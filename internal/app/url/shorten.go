package url

func (s *service) ShortenURL() {
	// var req UrlRequest
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	ctx.JSON(422, gin.H{
	// 		"error": "Validation error",
	// 	})
	// 	return
	// }

	// // Validate URL format
	// u, err := url.ParseRequestURI(req.Url)
	// if err != nil || u.Scheme == "" || u.Host == "" {
	// 	ctx.JSON(400, gin.H{
	// 		"error": "Invalid URL format",
	// 	})
	// 	return
	// }

	// // Check if URL length is valid (greater than 25 characters)
	// if len(req.Url) <= 25 {
	// 	ctx.JSON(400, gin.H{
	// 		"error": "URL must be longer than 25 characters",
	// 	})
	// 	return
	// }

	// // Create the short URL using the service
	// short, expireAt, err := s.UrlService.CreateShortUrl(req.Url, req.ExpireAt, ctx.Request.UserAgent())
	// if err != nil {
	// 	ctx.JSON(500, gin.H{
	// 		"error": "Internal server error",
	// 	})
	// 	return
	// }

	// // Build the full short URL with prefix
	// prefix := os.Getenv("SHORT_URL_PREFIX")
	// ctx.JSON(201, gin.H{
	// 	"message":   "success",
	// 	"short_url": prefix + short,
	// 	"expire_at": expireAt,
	// })
}
