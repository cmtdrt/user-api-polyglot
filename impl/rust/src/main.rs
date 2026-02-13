use axum::{routing::get, Router};

#[tokio::main]
async fn main() {
    let app = create_app();
    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080")
        .await
        .expect("Failed to bind to port 8080");

    println!("Server is running on http://localhost:8080");

    axum::serve(listener, app)
        .await
        .expect("Failed to start server");
}

fn create_app() -> Router {
    Router::new().route("/", get(hello_handler))
}

async fn hello_handler() -> &'static str {
    "Hello world from Rust!"
}