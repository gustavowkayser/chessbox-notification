package device

import (
	"chessbox-notification/internal/middleware"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var ErrDeviceNotFound = errors.New("device not found")

type Service struct {
	repository  Repository
	argonConfig ArgonConfig
}

func NewService(repository Repository, argonConfig ArgonConfig) *Service {
	return &Service{
		repository:  repository,
		argonConfig: argonConfig,
	}
}

func (s *Service) RegisterDevice(ctx context.Context, req RegisterDeviceRequest) (*Device, error) {
	id := uuid.NewString()

	credential := id + req.PushToken + req.AppVersion + req.Platform

	hash := sha256.Sum256([]byte(credential))
	credentialHash := hex.EncodeToString(hash[:])

	device := Device{
		ID:             id,
		CredentialHash: credentialHash,
		PushToken:      req.PushToken,
		AppVersion:     req.AppVersion,
		Platform:       req.Platform,
	}

	err := s.repository.Create(ctx, &device)

	if err != nil {
		return nil, err
	}

	return &device, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*middleware.Device, error) {
	device, err := s.repository.FindByCredentialsHash(ctx, s.hashPassword(token))

	if err != nil {
		return nil, err
	}

	if device == nil {
		return nil, ErrDeviceNotFound
	}

	d := middleware.Device{
		ID:             device.ID,
		CredentialHash: device.CredentialHash,
		PushToken:      device.PushToken,
		Platform:       device.Platform,
		AppVersion:     device.AppVersion,
	}

	return &d, nil
}

func (s *Service) UpdateDevice(ctx context.Context, req UpdateDeviceRequest) (*Device, error) {
	device, err := s.repository.FindByCredentialsHash(ctx, req.DeviceToken)

	if err != nil {
		return nil, err
	}

	device.PushToken = req.PushToken

	err = s.repository.Update(ctx, device.ID, device)

	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Service) UpsertDevice(
	ctx context.Context,
	req UpsertDeviceRequest,
	credentials *string,
) (*Device, error) {

	if credentials != nil {
		encodedHash := s.hashPassword(*credentials)
		device, err := s.repository.FindByCredentialsHash(ctx, encodedHash)

		if err != nil {
			return nil, err
		}

		if device != nil {
			device.PushToken = req.PushToken
			device.AppVersion = req.AppVersion
			device.Platform = req.Platform
			device.Active = req.Active
			device.Credentials = *credentials

			err := s.repository.Update(ctx, device.ID, device)

			if err != nil {
				return nil, err
			}

			return device, nil
		}
	}

	c := uuid.NewString()
	encodedHash := s.hashPassword(c)

	device := &Device{
		ID:             uuid.NewString(),
		Credentials:    c,
		CredentialHash: encodedHash,
		PushToken:      req.PushToken,
		Platform:       req.Platform,
		AppVersion:     req.AppVersion,
		Active:         req.Active,
	}

	err := s.repository.Create(ctx, device)

	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Service) hashPassword(password string) string {
	hash := argon2.IDKey(
		[]byte(password),
		s.argonConfig.Salt,
		s.argonConfig.TimeCost,
		s.argonConfig.MemoryCost,
		s.argonConfig.Threads,
		s.argonConfig.KeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.argonConfig.MemoryCost,
		s.argonConfig.TimeCost,
		s.argonConfig.Threads,
		base64.RawStdEncoding.EncodeToString(s.argonConfig.Salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}
