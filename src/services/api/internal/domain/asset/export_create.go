package asset

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	"github.com/qvora/api/internal/domain/identity"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// AllowedExportFormats — only a direct passthrough of the already-processed
// master asset is supported today. 4K upscale and GIF conversion both need a
// new transcode pass in the Rust postprocessor that doesn't exist yet; they
// deliberately return a clear "not supported" error rather than silently
// serving the wrong format.
var AllowedExportFormats = map[string]struct{}{
	"mp4_1080p": {},
}

// AllowedExportPlatforms — recorded on the export for naming/history purposes.
// All three currently produce the same 9:16 master asset: TikTok and YouTube
// Shorts already match that aspect ratio, and "meta" is accepted but does NOT
// yet trigger a 1:1 reframe (the Rust postprocessor's aspect_ratio parameter
// exists for this, but nothing wires a re-render pass into this endpoint yet
// — see the follow-up note in the handler below).
var AllowedExportPlatforms = map[string]struct{}{
	"meta":           {},
	"tiktok":         {},
	"youtube_shorts": {},
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "untitled"
	}
	return s
}

func newR2Client(ctx context.Context) (*s3.Client, string, error) {
	bucket := strings.TrimSpace(os.Getenv("R2_BUCKET"))
	if bucket == "" {
		return nil, "", fmt.Errorf("R2_BUCKET not set")
	}
	endpoint := strings.TrimSpace(os.Getenv("R2_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY"))
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, "", fmt.Errorf("R2_ENDPOINT, R2_ACCESS_KEY_ID, and R2_SECRET_ACCESS_KEY are required")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, "", fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
	return client, bucket, nil
}

type exportVariantRef struct {
	id           pgtype.UUID
	angle        string
	r2Key        string
	durationSecs *string
}

// CreateExports handles POST /api/v1/exports
// Bundles one or more already-processed variants into a single ZIP (with a
// manifest.csv), uploads it to R2, and records one `exports` row per variant
// pointing at the shared ZIP. Download happens via GET /api/v1/exports/:id/download,
// which mints a fresh short-lived presigned URL and tracks download_count —
// the raw R2 URL is never handed back directly from this endpoint.
func CreateExports(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		VariantIDs []string `json:"variant_ids"`
		Format     string   `json:"format"`
		Platform   string   `json:"platform"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	if len(req.VariantIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "variant_ids_required"})
	}
	if len(req.VariantIDs) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "max_100_variants_per_export"})
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "mp4_1080p"
	}
	if _, ok := AllowedExportFormats[format]; !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "format_not_supported",
			"message": "only mp4_1080p is supported today — 4K and GIF export need a Rust transcode pass that doesn't exist yet",
		})
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if platform == "" {
		platform = "tiktok"
	}
	if _, ok := AllowedExportPlatforms[platform]; !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}

	q, err := appmiddleware.GetScopedQueries(c)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	conn, err := appmiddleware.GetScopedConn(c)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	workspaceID, err := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	brandName, err := identity.GetBrandName(c, workspaceID)
	if err != nil {
		brandName = ""
	}
	if brandName == "" {
		brandName = "brand"
	}

	refs := make([]exportVariantRef, 0, len(req.VariantIDs))
	for _, rawID := range req.VariantIDs {
		variantUUID, err := util.ParseUUID(strings.TrimSpace(rawID))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid_variant_id: %s", rawID)})
		}
		variant, err := q.GetVariantForPlayback(c.Request().Context(), db.GetVariantForPlaybackParams{
			ID:          variantUUID,
			WorkspaceID: workspaceID,
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("variant_not_found: %s", rawID)})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_lookup_failed"})
		}
		if !strings.EqualFold(variant.Status, "complete") {
			return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("variant_not_complete: %s", rawID)})
		}
		if variant.R2Key == nil || strings.TrimSpace(*variant.R2Key) == "" {
			return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("variant_missing_asset: %s", rawID)})
		}
		var durationStr *string
		if variant.DurationSecs.Valid {
			if f, err := variant.DurationSecs.Float64Value(); err == nil && f.Valid {
				s := fmt.Sprintf("%.2f", f.Float64)
				durationStr = &s
			}
		}
		refs = append(refs, exportVariantRef{
			id:           variant.ID,
			angle:        variant.Angle,
			r2Key:        *variant.R2Key,
			durationSecs: durationStr,
		})
	}

	client, bucket, err := newR2Client(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "r2_not_configured"})
	}

	zipBuf := &bytes.Buffer{}
	zw := zip.NewWriter(zipBuf)
	manifestBuf := &bytes.Buffer{}
	manifestWriter := csv.NewWriter(manifestBuf)
	if err := manifestWriter.Write([]string{"filename", "variant_id", "angle", "platform", "duration_secs"}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "manifest_write_failed"})
	}

	today := time.Now().UTC().Format("20060102")
	brandSlug := slugify(brandName)

	for i, ref := range refs {
		obj, err := client.GetObject(c.Request().Context(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(ref.r2Key),
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("asset_download_failed: %s", util.UUIDString(ref.id))})
		}
		filename := fmt.Sprintf("%s_%s_%s_%s_v%d.mp4", brandSlug, slugify(ref.angle), platform, today, i+1)
		writer, err := zw.Create(filename)
		if err != nil {
			obj.Body.Close() //nolint:errcheck
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "zip_write_failed"})
		}
		if _, err := io.Copy(writer, obj.Body); err != nil {
			obj.Body.Close() //nolint:errcheck
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "zip_write_failed"})
		}
		obj.Body.Close() //nolint:errcheck

		duration := ""
		if ref.durationSecs != nil {
			duration = *ref.durationSecs
		}
		if err := manifestWriter.Write([]string{filename, util.UUIDString(ref.id), ref.angle, platform, duration}); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "manifest_write_failed"})
		}
	}

	manifestWriter.Flush()
	if err := manifestWriter.Error(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "manifest_write_failed"})
	}
	manifestEntry, err := zw.Create("manifest.csv")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "zip_write_failed"})
	}
	if _, err := manifestEntry.Write(manifestBuf.Bytes()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "zip_write_failed"})
	}
	if err := zw.Close(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "zip_finalize_failed"})
	}

	zipKey := fmt.Sprintf("exports/%s/%s.zip", util.UUIDString(workspaceID), uuid.NewString())
	if _, err := client.PutObject(c.Request().Context(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(zipKey),
		Body:        bytes.NewReader(zipBuf.Bytes()),
		ContentType: aws.String("application/zip"),
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "zip_upload_failed"})
	}

	exportIDs := make([]string, 0, len(refs))
	for _, ref := range refs {
		row := conn.QueryRow(
			c.Request().Context(),
			`INSERT INTO exports (workspace_id, variant_id, destination, status, r2_key)
			 VALUES ($1, $2, $3, 'complete', $4)
			 RETURNING id`,
			workspaceID, ref.id, platform, zipKey,
		)
		var exportID pgtype.UUID
		if err := row.Scan(&exportID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "export_record_failed"})
		}
		exportIDs = append(exportIDs, util.UUIDString(exportID))
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"org_id":        claims.OrgID,
		"export_ids":    exportIDs,
		"variant_count": len(refs),
		"format":        format,
		"platform":      platform,
		"download_url":  fmt.Sprintf("/api/v1/exports/%s/download", exportIDs[0]),
	})
}

// DownloadExport handles GET /api/v1/exports/:id/download
// Mints a fresh 15-minute presigned URL for the export's ZIP, increments
// download_count, and redirects. The raw R2 URL is never returned from
// CreateExports directly so every download goes through this counter.
func DownloadExport(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	exportID, err := util.ParseUUID(strings.TrimSpace(c.Param("id")))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_export_id"})
	}
	workspaceID, err := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	conn, err := appmiddleware.GetScopedConn(c)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	var r2Key *string
	row := conn.QueryRow(
		c.Request().Context(),
		`SELECT r2_key FROM exports WHERE id = $1 AND workspace_id = $2`,
		exportID, workspaceID,
	)
	if err := row.Scan(&r2Key); err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "export_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "export_lookup_failed"})
	}
	if r2Key == nil || strings.TrimSpace(*r2Key) == "" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "export_not_ready"})
	}

	client, bucket, err := newR2Client(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "r2_not_configured"})
	}
	presignClient := s3.NewPresignClient(client)
	presigned, err := presignClient.PresignGetObject(c.Request().Context(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(*r2Key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "presign_failed"})
	}

	_, _ = conn.Exec(
		c.Request().Context(),
		`UPDATE exports SET download_count = download_count + 1, updated_at = NOW() WHERE id = $1 AND workspace_id = $2`,
		exportID, workspaceID,
	)

	return c.Redirect(http.StatusFound, presigned.URL)
}
