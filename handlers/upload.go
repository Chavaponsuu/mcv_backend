package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	
	"mcv_backend/services"
	

)
type GenerateSignatureRequest struct {
	Folder   string `json:"folder" binding:"required"`
	PublicID string `json:"public_id" binding:"required"`
	Context string `json:"context" binding:"required"`
	Tags string `json:"tags" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
}

func GenerateSignatureHandler(w http.ResponseWriter, r *http.Request) {
	var req GenerateSignatureRequest

	// parse JSON
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// timestamp := time.Now().Unix()

	params := map[string]string{
	"context":   req.Context,
	"folder":    req.Folder,
	"public_id": req.PublicID, // ✅ ต้องมีแล้ว!
	"tags":      req.Tags,
	"timestamp": req.Timestamp,
}

	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	signature := services.GenerateSignature(params, apiSecret)

	// set header
	w.Header().Set("Content-Type", "application/json")

	// response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"signature": signature,
		"timestamp": req.Timestamp,
	})
}