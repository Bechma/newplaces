use rocket::fs::FileServer;
use rocket::http::{ContentType, Header};
use rocket::log::private::{log, Level};
use rocket::response::content::RawHtml;
use rocket::response::stream::{Event, EventStream};
use rocket::serde::json::Json;
use rocket::serde::{Deserialize, Serialize};
use rocket::shield::{Policy, Shield};
use rocket::tokio::fs::File;
use rocket::tokio::select;
use rocket::tokio::sync::broadcast::{channel, error::RecvError, Sender};
use rocket::tokio::sync::RwLock;
use rocket::{get, post};
use rocket::{routes, Shutdown, State};
use rocket_db_pools::deadpool_redis::redis;
use rocket_db_pools::Database;

const CANVAS_NAME: &str = "newplaces";

#[derive(Database)]
#[database("redis_db")]
struct Redis(rocket_db_pools::deadpool_redis::Pool);

#[derive(Serialize, Deserialize, Clone)]
#[serde(crate = "rocket::serde")]
struct Pixel {
    pub x: u32,
    pub y: u32,
    pub color: u32,
}

#[get("/")]
async fn index() -> Option<RawHtml<File>> {
    File::open("./ui/dist/index.html").await.map(RawHtml).ok()
}

#[get("/events")]
async fn events(queue: &State<Sender<Pixel>>, mut end: Shutdown) -> EventStream![] {
    let mut rx = queue.subscribe();
    EventStream! {
        loop {
            let msg = select! {
                msg = rx.recv() => match msg {
                    Ok(msg) => msg,
                    Err(RecvError::Closed) => break,
                    Err(RecvError::Lagged(_)) => continue,
                },
                _ = &mut end => break,
            };

            yield Event::json(&msg);
        }
    }
}

#[post("/pixel", format = "json", data = "<px>")]
async fn pixel(
    px: Json<Pixel>,
    mut db: rocket_db_pools::Connection<Redis>,
    queue: &State<Sender<Pixel>>,
    canvas: &State<RwLock<Vec<u8>>>,
) -> rocket::http::Status {
    if px.x >= 2000 || px.x >= 2000 {
        return rocket::http::Status::BadRequest;
    }
    let position = (px.y * 2000 + px.x) as usize;
    match redis::cmd("BITFIELD")
        .arg(CANVAS_NAME)
        .arg("SET")
        .arg("u32")
        .arg(32 * position)
        .arg(px.color)
        .query_async::<_, ()>(&mut *db)
        .await
    {
        Err(err) => {
            log!(Level::Error, "Redis error {}", err);
            rocket::http::Status::TooManyRequests
        }
        _ => {
            let position = position * 4;
            let mut w = canvas.inner().write().await;
            w[position] = (px.color >> 24 & 0xFF) as u8;
            w[position + 1] = (px.color >> 16 & 0xFF) as u8;
            w[position + 2] = (px.color >> 8 & 0xFF) as u8;
            w[position + 3] = (px.color & 0xFF) as u8;
            // A send 'fails' if there are no active subscribers. That's okay.
            let _ = queue.send(px.0);
            rocket::http::Status::Ok
        }
    }
}

#[get("/canvas", format = "binary")]
async fn canvas(canvas: &State<RwLock<Vec<u8>>>) -> (ContentType, Vec<u8>) {
    let c = canvas.inner().read().await;
    (ContentType::Binary, c.to_vec())
}

// Access-Control-Allow-Origin
#[derive(Default)]
struct AccessControlAllowOrigin;

impl Policy for AccessControlAllowOrigin {
    const NAME: &'static str = "Access-Control-Allow-Origin";

    fn header(&self) -> Header<'static> {
        Header::new(Self::NAME, "*")
    }
}

async fn get_canvas(rocket_build: &rocket::Rocket<rocket::Build>) -> Vec<u8> {
    let res = rocket_build.figment();
    let redis_url = redis::parse_redis_url(
        res.find_value("databases.redis_db.url")
            .expect("redis url not found")
            .as_str()
            .expect("bad redis url"),
    )
    .expect("bad redis url");
    let mut conn = redis::Client::open(redis_url)
        .unwrap()
        .get_async_connection()
        .await
        .unwrap();
    redis::cmd("GET")
        .arg(CANVAS_NAME)
        .query_async(&mut conn)
        .await
        .unwrap()
}

#[rocket::main]
async fn main() {
    let rocket_build = rocket::build();
    let canvas = RwLock::new(get_canvas(&rocket_build).await);
    let shield = Shield::default().enable(AccessControlAllowOrigin);
    let _ = rocket_build
        .attach(Redis::init())
        .attach(shield)
        .manage(channel::<Pixel>(1024).0)
        .manage(canvas)
        .mount("/", routes![index, events, pixel, canvas])
        .mount("/", FileServer::from("ui/dist"))
        .launch()
        .await;
}
