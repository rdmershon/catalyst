package catalyst

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"

	"github.com/SecurityBrewery/catalyst/database"
	"github.com/SecurityBrewery/catalyst/pointer"
	"github.com/SecurityBrewery/catalyst/storage"
)

func RestoreHandler(catalystStorage *storage.Storage, db *database.Database, c *database.Config) gin.HandlerFunc {
	return func(context *gin.Context) {
		uf, err := context.FormFile("backup")
		if err != nil {
			log.Println(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		f, err := uf.Open()
		if err != nil {
			log.Println(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		defer f.Close()

		if err = Restore(context, catalystStorage, db, c, f, uf.Size); err != nil {
			log.Println(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func Restore(ctx context.Context, catalystStorage *storage.Storage, db *database.Database, c *database.Config, r io.Reader, size int64) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	ra := bytes.NewReader(b)
	fsys, err := zip.NewReader(ra, size)
	if err != nil {
		return err
	}

	if fsys.Comment != GetVersion() {
		return errors.New(fmt.Sprintf("wrong version, got: %s, want: %s", fsys.Comment, GetVersion()))
	}

	dir, err := os.MkdirTemp("", "catalyst-restore")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err = unzip(fsys, dir); err != nil {
		return err
	}

	if err := restoreS3(catalystStorage, path.Join(dir, "minio")); err != nil {
		return err
	}

	if err := arangorestore(path.Join(dir, "arango"), c); err != nil {
		return err
	}

	return db.IndexRebuild(ctx)
}

func restoreS3(catalystStorage *storage.Storage, p string) error {
	minioDir := os.DirFS(p)

	entries, err := fs.ReadDir(minioDir, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := restoreBucket(catalystStorage, entry, minioDir); err != nil {
			return err
		}
	}
	return nil
}

func restoreBucket(catalystStorage *storage.Storage, entry fs.DirEntry, minioDir fs.FS) error {
	_, err := catalystStorage.S3().CreateBucket(&s3.CreateBucketInput{Bucket: pointer.String(entry.Name())})
	if err != nil {
		return err
	}

	uploader := catalystStorage.Uploader()

	f, err := minioDir.Open(entry.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	err = fs.WalkDir(minioDir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		_, err = uploader.Upload(&s3manager.UploadInput{Body: f, Bucket: pointer.String(entry.Name()), Key: pointer.String(path)})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func unzip(archive *zip.Reader, dir string) error {
	return fs.WalkDir(archive, "arango", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			_ = os.MkdirAll(path.Join(dir, p), os.ModePerm)
			return nil
		}

		f, err := archive.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		return os.WriteFile(path.Join(dir, p), b, os.ModePerm)
	})
}

func arangorestore(dir string, config *database.Config) error {
	host := strings.Replace(config.Host, "http", "tcp", 1)

	name := config.Name
	if config.Name == "" {
		name = database.Name
	}
	args := []string{
		"--input-directory", dir, "--server.endpoint", host,
		"--server.username", config.User, "--server.password", config.Password,
		"--server.database", name,
	}
	cmd := exec.Command("arangorestore", args...)
	return cmd.Run()
}
