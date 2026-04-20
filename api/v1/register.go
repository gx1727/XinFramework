package v1

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/pkg/config"
	jwtpkg "gx1727.com/xin/pkg/jwt"
	"gx1727.com/xin/pkg/resp"
)

type loginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID uint   `json:"tenant_id"`
}

type accountRow struct {
	ID       uint
	Username string
	Phone    string
	Email    string
	Password string
}

type userRow struct {
	ID       uint
	TenantID uint
	Code     string
	Status   int16
}

func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			resp.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/login", func(c *gin.Context) {
			var req loginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				resp.BadRequest(c, "invalid request body")
				return
			}

			d := db.Get()
			if d == nil {
				resp.ServerError(c, "database not initialized")
				return
			}

			var acc accountRow
			if err := d.Table("accounts").
				Select("id, username, phone, email, password").
				Where("is_deleted = FALSE").
				Where("username = ? OR phone = ? OR email = ?", req.Account, req.Account, req.Account).
				First(&acc).Error; err != nil {
				resp.Unauthorized(c, "invalid account or password")
				return
			}

			ok, err := verifyPassword(acc.Password, req.Password)
			if err != nil || !ok {
				resp.Unauthorized(c, "invalid account or password")
				return
			}

			q := d.Table("users").
				Select("id, tenant_id, code, status").
				Where("is_deleted = FALSE").
				Where("account_id = ?", acc.ID)
			if req.TenantID > 0 {
				q = q.Where("tenant_id = ?", req.TenantID)
			}

			var u userRow
			if err := q.Order("id ASC").First(&u).Error; err != nil {
				resp.Forbidden(c, "user is not bound to any tenant")
				return
			}
			if u.Status != 1 {
				resp.Forbidden(c, "user is disabled")
				return
			}

			roleCode := "user"
			var role struct {
				Code string
			}
			_ = d.Table("user_roles ur").
				Select("r.code").
				Joins("JOIN roles r ON r.id = ur.role_id").
				Where("ur.is_deleted = FALSE").
				Where("r.is_deleted = FALSE").
				Where("ur.user_id = ?", u.ID).
				Order("ur.id ASC").
				First(&role).Error
			if role.Code != "" {
				roleCode = role.Code
			}

			cfg := config.Get()
			if cfg == nil {
				resp.ServerError(c, "config not initialized")
				return
			}

			token, err := jwtpkg.Generate(&cfg.JWT, u.ID, u.TenantID, roleCode)
			if err != nil {
				resp.ServerError(c, "generate token failed")
				return
			}

			resp.Success(c, gin.H{
				"token": token,
				"user": gin.H{
					"id":        u.ID,
					"tenant_id": u.TenantID,
					"code":      u.Code,
					"role":      roleCode,
				},
				"session_id": uuid.NewString(),
			})
		})
	}
}

func verifyPassword(stored, plain string) (bool, error) {
	// 兼容开发环境明文密码
	if !strings.HasPrefix(stored, "$argon2id$") {
		return subtle.ConstantTimeCompare([]byte(stored), []byte(plain)) == 1, nil
	}
	return verifyArgon2ID(stored, plain)
}

func verifyArgon2ID(encodedHash, plain string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid argon2id hash format")
	}
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported hash algorithm")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLength := uint32(len(hash))
	calculated := argon2.IDKey([]byte(plain), salt, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(hash, calculated) == 1, nil
}
