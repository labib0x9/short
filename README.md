# PikachuAPI - A Url Shortener

A simple URL shortener service written in Go, using Gin for HTTP routing, MySQL for persistent storage, and Redis for caching.

## How to Use
1. **Shorten a URL**
   - Send a `POST` request to `/shorten` with JSON body:
     ```json
     {
       "url": "https://example.com",
       "expire_in": 60 // (optional) expiration in minutes
     }
     ```
   - Response:
     ```json
     {
       "message": "success",
       "short_url": "http://localhost:8080/IrLvWOeO",
       "expire_at" : "2025-05-12 12:23:06"
     }
     ```
2. **Redirect to Original URL**
   - Access `GET /:code` (e.g., `/IrLvWOeO`)
   - If the code exists and is not expired, you will be redirected to the original URL.
3. **Fetch metadata of short URL**
   - Access `GET /fetch/:code` (e.g., `/fetch/IrLvWOeO`)
   - If the code exists , you will see the Metadata of the short URL.


# MySql Setup
**Create Database**
```sql
CREATE DATABASE urlshortener;
```
**Create Table**
```sql
USE urlshortener;

CREATE TABLE urls (
    id INT AUTO_INCREMENT PRIMARY KEY,
    url TEXT NOT NULL,
    short_url VARCHAR(64) NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    expire DATETIME
);
```
