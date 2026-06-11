package url

import "net/http"

func (h *Handler) Analysis(w http.ResponseWriter, r *http.Request) {
}

// shortCode := ctx.Param("code")
// if shortCode == "" {
// 	ctx.JSON(400, gin.H{
// 		"error": "Short URL code is required",
// 	})
// 	return
// }
// url, err := s.UrlService.GetUrlByCode(shortCode)

// if err != nil && err != utils.ErrUrlNotFound {
// 	errMsg := "Internal server error"
// 	errCode := 500
// 	switch err {
// 	case utils.ErrUrlNotFound:
// 		errMsg, errCode = "URL not found", 404

// 		// case utils.ErrShortCodeExpired:
// 		// 	errMsg, errCode = "URL has expired", 410
// 	}

// 	ctx.JSON(errCode, gin.H{
// 		"error": errMsg,
// 	})
// 	return
// }

// ctx.JSON(200, gin.H{
// 	"url": url.URL,
// 	"metadata": gin.H{
// 		"short_code": shortCode,
// 		"created_at": url.CreatedAt,
// 		"expire_at":  url.Expire,
// 	},
// })
// }
