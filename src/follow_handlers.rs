use actix_web::{web, HttpResponse};
use serde::Deserialize;
use aws_sdk_dynamodb::Client; // Import the DynamoDB client
use crate::follow_db;
use tracing::{info, error};


#[derive(Debug, Deserialize)]
pub struct FollowRequest {
    user_id: String, // ID of the user to follow or unfollow
}

// POST /users/{user_id}/follow
pub async fn follow_user(
    client: web::Data<Client>, // Access the DynamoDB client from the app state
    path: web::Path<String>, 
    req: web::Json<FollowRequest>
) -> HttpResponse {
    let follower_id = path.into_inner();
    let followed_id = req.user_id.clone();

    match follow_db::add_follow(&client, &follower_id, &followed_id).await {
        Ok(_) => HttpResponse::Ok().body("Followed successfully"),
        
        Err(err) => {
            error!("Failed to add follower :{:?}", err);
            return HttpResponse::InternalServerError().body(err.to_string());
        },
    }
}

// POST /users/{user_id}/unfollow
pub async fn unfollow_user(
    client: web::Data<Client>, // Access the DynamoDB client from the app state
    path: web::Path<String>, 
    req: web::Json<FollowRequest>
) -> HttpResponse {
    let follower_id = path.into_inner();
    let followed_id = req.user_id.clone();

    match follow_db::remove_follow(&client, &follower_id, &followed_id).await {
        Ok(_) => HttpResponse::Ok().body("Unfollowed successfully"),
        Err(err) => HttpResponse::InternalServerError().body(err.to_string()),
    }
}

// GET /users/{user_id}/follows
pub async fn get_user_follows(
    client: web::Data<Client>, // Access the DynamoDB client from the app state
    path: web::Path<String>
) -> HttpResponse {
    let follower_id = path.into_inner();

    match follow_db::get_follows(&client, &follower_id).await {
        Ok(follows) => HttpResponse::Ok().json(follows),
        Err(err) => {
            error!("Failed to add follower :{:?}", err);
            return HttpResponse::InternalServerError().body(err.to_string());
        },
    }
}

// POST /users/{user_id}/delete_follower
pub async fn delete_follower(
    client: web::Data<Client>, // Access the DynamoDB client from the app state
    path: web::Path<String>, 
    req: web::Json<FollowRequest>
) -> HttpResponse {
    let followed_id = path.into_inner(); // This is the user who is followed
    let follower_id = req.user_id.clone(); // This is the follower to be removed

    // Match the result of removing the follow relationship
    match follow_db::remove_follow(&client, &follower_id, &followed_id).await {
        Ok(_) => HttpResponse::Ok().body("Follower removed successfully"),
        Err(err) => HttpResponse::InternalServerError().body(err.to_string()),
    }
}

// GET /users/{user_id}/followers
pub async fn get_followers(
    client: web::Data<Client>, // Access the DynamoDB client from the app state
    path: web::Path<String>
) -> HttpResponse {
    let followed_id = path.into_inner();

    // Query the Follow table to find all users following the given `followed_id`
    match follow_db::get_follows(&client, &followed_id).await {
        Ok(followers) => HttpResponse::Ok().json(followers),
        Err(err) => {
            error!("Failed to add follower :{:?}", err);
            return HttpResponse::InternalServerError().body(err.to_string());
        },
    }
}