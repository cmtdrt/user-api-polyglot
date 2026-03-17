use std::net::SocketAddr;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    routing::{delete, get, post, put},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use sqlx::{postgres::PgPoolOptions, PgPool};
use uuid::Uuid;

#[derive(Clone)]
struct AppState {
    pool: PgPool,
}

#[derive(Debug, Serialize)]
struct User {
    id: Uuid,
    name: String,
    email: String,
    created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
struct NewUser {
    name: String,
    email: String,
}

#[derive(Debug, Deserialize)]
struct UpdateUser {
    name: Option<String>,
    email: Option<String>,
}

#[tokio::main]
async fn main() -> Result<(), anyhow::Error> {
    let database_url =
        std::env::var("DATABASE_URL").expect("DATABASE_URL env var must be set for the API");

    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect(&database_url)
        .await?;

    let state = AppState { pool };

    let app = create_app(state);

    let addr: SocketAddr = "0.0.0.0:8080".parse().unwrap();
    println!("Rust API listening on http://{}", addr);

    axum::serve(tokio::net::TcpListener::bind(addr).await?, app).await?;

    Ok(())
}

fn create_app(state: AppState) -> Router {
    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route(
            "/users/:id",
            get(get_user).put(update_user).delete(delete_user),
        )
        .with_state(state)
}

async fn list_users(State(state): State<AppState>) -> Result<Json<Vec<User>>, StatusCode> {
    let users = sqlx::query_as!(
        User,
        r#"SELECT id, name, email, created_at
           FROM users
           ORDER BY created_at DESC"#
    )
    .fetch_all(&state.pool)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(users))
}

async fn get_user(State(state): State<AppState>, Path(id): Path<Uuid>) -> Result<Json<User>, StatusCode> {
    let user = sqlx::query_as!(
        User,
        r#"SELECT id, name, email, created_at
           FROM users
           WHERE id = $1"#,
        id
    )
    .fetch_optional(&state.pool)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    match user {
        Some(u) => Ok(Json(u)),
        None => Err(StatusCode::NOT_FOUND),
    }
}

async fn create_user(State(state): State<AppState>, Json(payload): Json<NewUser>) -> Result<(StatusCode, Json<User>), StatusCode> {
    let user = sqlx::query_as!(
        User,
        r#"INSERT INTO users (name, email)
           VALUES ($1, $2)
           RETURNING id, name, email, created_at"#,
        payload.name,
        payload.email,
    )
    .fetch_one(&state.pool)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(user)))
}

async fn update_user(State(state): State<AppState>, Path(id): Path<Uuid>, Json(payload): Json<UpdateUser>) -> Result<Json<User>, StatusCode> {
    let current = sqlx::query_as!(
        User,
        r#"SELECT id, name, email, created_at
           FROM users
           WHERE id = $1"#,
        id
    )
    .fetch_optional(&state.pool)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let Some(current) = current else {
        return Err(StatusCode::NOT_FOUND);
    };

    let new_name = payload.name.unwrap_or(current.name);
    let new_email = payload.email.unwrap_or(current.email);

    let user = sqlx::query_as!(
        User,
        r#"UPDATE users
           SET name = $1, email = $2
           WHERE id = $3
           RETURNING id, name, email, created_at"#,
        new_name,
        new_email,
        id
    )
    .fetch_one(&state.pool)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(user))
}

async fn delete_user(State(state): State<AppState>, Path(id): Path<Uuid>) -> Result<StatusCode, StatusCode> {
    let result = sqlx::query!(r#"DELETE FROM users WHERE id = $1"#, id)
        .execute(&state.pool)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    if result.rows_affected() == 0 {
        return Err(StatusCode::NOT_FOUND);
    }

    Ok(StatusCode::NO_CONTENT)
}