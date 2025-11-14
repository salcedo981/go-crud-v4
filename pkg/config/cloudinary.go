package config

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	utils_v1 "github.com/FDSAP-Git-Org/hephaestus/utils/v1"
)

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
	Folder    string
}

func LoadCloudinaryConfig() CloudinaryConfig {
	return CloudinaryConfig{
		CloudName: utils_v1.GetEnv("CLOUDINARY_CLOUD_NAME"),
		APIKey:    utils_v1.GetEnv("CLOUDINARY_API_KEY"),
		APISecret: utils_v1.GetEnv("CLOUDINARY_API_SECRET"),
		Folder:    "expenses/uploads",
	}
}

func UploadToCloudinary(fileHeader *multipart.FileHeader, config CloudinaryConfig) (string, error) {
	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		return "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	randomStr := utils_v1.GenerateRandomStrings(8, []string{utils_v1.UpperString, utils_v1.LowerString, utils_v1.NumericString})

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID: config.Folder + "/" + randomStr,
		Folder:   config.Folder,
	})
	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

// Delete an image from Cloudinary using PublicID
func DeleteCloudinaryImage(publicID string, config CloudinaryConfig) error {
	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	return err
}

// Extract PublicID from Cloudinary URL
func CloudinaryPublicIDFromURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}

	parts := strings.Split(imageURL, "/")
	idx := -1
	for i, p := range parts {
		if p == "upload" {
			idx = i
			break
		}
	}
	if idx == -1 || idx+2 >= len(parts) {
		return ""
	}

	pathParts := parts[idx+2:] // skip "upload" + version
	path := strings.Join(pathParts, "/")
	return strings.TrimSuffix(path, filepath.Ext(path))
}
