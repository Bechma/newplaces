use std::collections::HashSet;
use std::time::Duration;

use actix::prelude::*;
use actix_broker::BrokerSubscribe;

use futures_util::FutureExt;

use crate::message::{Connection, Subscriber};
use crate::WsCanvasSession;

#[derive(Default)]
pub struct RedisServer {
    client: Option<redis::Client>,
    sessions: HashSet<Addr<WsCanvasSession>>,
    connection: Option<redis::aio::MultiplexedConnection>,
}

impl RedisServer {
    #[inline]
    fn get_connection(&self) -> redis::aio::MultiplexedConnection {
        self.connection.as_ref().unwrap().clone()
    }
}

impl Actor for RedisServer {
    type Context = Context<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        self.subscribe_system_async::<Subscriber>(ctx);
    }
}

impl Handler<Subscriber> for RedisServer {
    type Result = ();

    fn handle(&mut self, msg: Subscriber, ctx: &mut Self::Context) -> Self::Result {
        let mut conn = self.get_connection();
        let message = msg.clone();
        ctx.spawn(
            fut::wrap_future::<_, Self>(async move { msg.run_query(&mut conn).await }).map(
                move |x, act, ctx| {
                    match x {
                        // inform the others
                        Ok(_) => act.sessions.iter().for_each(|x| x.do_send(message.clone())),
                        Err(e) => {
                            log::error!("Redis error {}", e);
                            ctx.stop();
                        }
                    }
                },
            ),
        );
    }
}

impl Handler<Connection> for RedisServer {
    type Result = ResponseFuture<Result<Vec<u8>, redis::RedisError>>;

    fn handle(&mut self, msg: Connection, _ctx: &mut Self::Context) -> Self::Result {
        if !self.sessions.remove(&msg.addr) {
            self.sessions.insert(msg.addr);
            let mut conn = self.get_connection();
            return async move {
                redis::cmd("GET")
                    .arg("newplaces")
                    .query_async::<redis::aio::MultiplexedConnection, Vec<u8>>(&mut conn)
                    .await
            }
            .boxed();
        }
        log::info!("Client disconnected, {} remaining", self.sessions.len());
        async { Ok(Vec::new()) }.boxed()
    }
}

impl Supervised for RedisServer {
    fn restarting(&mut self, ctx: &mut <Self as Actor>::Context) {
        let client = self.client.as_ref().unwrap().clone();
        ctx.wait(
            fut::wrap_future::<_, Self>(
                async move { client.get_multiplexed_async_connection().await },
            )
            .map(|conn, act, ctx| match conn {
                Ok(x) => {
                    log::info!("Connection recovered");
                    act.connection = Some(x);
                }
                Err(e) => {
                    ctx.run_later(Duration::from_secs(2), move |_, ctx| {
                        log::warn!("Restarted due to: {}", e);
                        ctx.stop();
                    });
                }
            }),
        );
    }
}

impl SystemService for RedisServer {
    fn service_started(&mut self, ctx: &mut Context<Self>) {
        ctx.wait(
            fut::wrap_future::<_, Self>(async {
                let client = redis::Client::open(std::env::var("REDIS_URN").unwrap()).unwrap();
                let connection = client.get_multiplexed_async_connection().await.unwrap();
                (client, connection)
            })
            .map(|(client, connection), act, _ctx| {
                act.client = Some(client);
                act.connection = Some(connection);
            }),
        )
    }
}
