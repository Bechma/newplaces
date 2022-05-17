use rocket::http::Status;
use rocket::{async_test, uri};

use newplaces::rocket_uri_macro_pixel;
use newplaces::{rocket_client, Pixel};

#[async_test]
async fn test_set_pixel_fail() {
    let client = rocket::local::asynchronous::Client::tracked(rocket_client().await.unwrap())
        .await
        .unwrap();
    let response = client
        .post(uri!(pixel))
        .json(&Pixel {
            x: 3000,
            y: 0,
            color: 0,
        })
        .dispatch()
        .await;
    assert_eq!(response.status(), Status::BadRequest);

    let response = client
        .post(uri!(pixel))
        .json(&Pixel {
            x: 0,
            y: 3000,
            color: 0,
        })
        .dispatch()
        .await;
    assert_eq!(response.status(), Status::BadRequest);
}

#[async_test]
async fn test_set_pixel_success() {
    let client = rocket::local::asynchronous::Client::tracked(rocket_client().await.unwrap())
        .await
        .unwrap();
    let response = client
        .post(uri!(pixel))
        .json(&Pixel {
            x: 0,
            y: 0,
            color: 0,
        })
        .dispatch()
        .await;
    assert_eq!(response.status(), Status::Ok)
}
