package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-banking-service/internal/dto"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
	"go-banking-service/internal/repository"
	"io"
	"math/rand"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"golang.org/x/crypto/bcrypt"
)

type CardService interface {
	CreateCard(ctx context.Context, userID string, req *dto.CreateCardRequest) (*dto.CardResponse, error)
	GetCardsByUserID(ctx context.Context, userID string) ([]*dto.CardResponse, error)
}

type cardService struct {
	cardRepo    repository.CardRepository
	accountRepo repository.AccountRepository
	hmacSecret  []byte

	publicEntity  *openpgp.Entity
	privateEntity *openpgp.Entity
}

func NewCardService(
	cardRepo repository.CardRepository,
	accountRepo repository.AccountRepository,
	hmacSecret string,
	publicKey string,
	privateKey string,
) CardService {

	pubList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(publicKey)))
	if err != nil || len(pubList) == 0 {
		logger.Fatal("Failed to parse PUBLIC key", "error", err)
	}

	privList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(privateKey)))
	if err != nil || len(privList) == 0 {
		logger.Fatal("Failed to parse PRIVATE key", "error", err)
	}

	return &cardService{
		cardRepo:      cardRepo,
		accountRepo:   accountRepo,
		hmacSecret:    []byte(hmacSecret),
		publicEntity:  pubList[0],
		privateEntity: privList[0],
	}
}

func (s *cardService) CreateCard(ctx context.Context, userID string, req *dto.CreateCardRequest) (*dto.CardResponse, error) {
	account, err := s.accountRepo.GetByIDAndUserID(ctx, req.AccountID, userID)
	if err != nil {
		logger.Error("Failed to get account for card creation", "error", err, "account_id", req.AccountID, "user_id", userID)
		return nil, err
	}

	cardNumber := generateCardNumber()
	expiryDate := generateExpiryDate()
	cvv := generateCVV()

	encryptedNumber, err := encryptDataPGP(cardNumber, s.publicEntity)
	if err != nil {
		logger.Error("Failed to encrypt card number", "error", err)
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}

	encryptedExpiry, err := encryptDataPGP(expiryDate, s.publicEntity)
	if err != nil {
		logger.Error("Failed to encrypt expiry date", "error", err)
		return nil, fmt.Errorf("failed to encrypt expiry date: %w", err)
	}

	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash CVV", "error", err)
		return nil, fmt.Errorf("failed to hash CVV: %w", err)
	}

	hmacValue := computeHMAC(cardNumber+expiryDate, s.hmacSecret)

	card := &model.Card{
		ID:              generateUUID(),
		UserID:          userID,
		AccountID:       account.ID,
		NumberEncrypted: encryptedNumber,
		ExpiryEncrypted: encryptedExpiry,
		CVVHash:         cvvHash,
		HMAC:            hmacValue,
		CreatedAt:       time.Now(),
	}

	err = s.cardRepo.Create(ctx, card)
	if err != nil {
		logger.Error("Failed to create card in database", "error", err, "card_id", card.ID)
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	lastFour := cardNumber[len(cardNumber)-4:]

	logger.Info("Card created successfully", "card_id", card.ID, "user_id", userID, "last_four", lastFour)

	return &dto.CardResponse{
		ID:             card.ID,
		UserID:         card.UserID,
		AccountID:      card.AccountID,
		NumberLastFour: lastFour,
		ExpiryDate:     expiryDate,
		CreatedAt:      card.CreatedAt,
	}, nil
}

func (s *cardService) GetCardsByUserID(ctx context.Context, userID string) ([]*dto.CardResponse, error) {
	cards, err := s.cardRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get cards for user", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get cards: %w", err)
	}

	var responses []*dto.CardResponse
	for _, card := range cards {
		decryptedNumber, err := decryptDataPGP(card.NumberEncrypted, s.privateEntity)
		if err != nil {
			logger.Error("Failed to decrypt card number", "error", err, "card_id", card.ID)
			continue
		}
		decryptedExpiry, err := decryptDataPGP(card.ExpiryEncrypted, s.privateEntity)
		if err != nil {
			logger.Error("Failed to decrypt expiry date", "error", err, "card_id", card.ID)
			continue
		}

		lastFour := decryptedNumber[len(decryptedNumber)-4:]

		responses = append(responses, &dto.CardResponse{
			ID:             card.ID,
			UserID:         card.UserID,
			AccountID:      card.AccountID,
			NumberLastFour: lastFour,
			ExpiryDate:     decryptedExpiry,
			CreatedAt:      card.CreatedAt,
		})
	}
	return responses, nil
}

func generateCardNumber() string {
	prefix := "4"
	length := 16
	number := prefix
	for i := 0; i < length-1; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}

	sum := 0
	isSecond := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		isSecond = !isSecond
	}
	checkDigit := (10 - (sum % 10)) % 10
	return number + fmt.Sprintf("%d", checkDigit)
}

func generateExpiryDate() string {
	now := time.Now()
	expiry := now.AddDate(3, 0, 0)
	return expiry.Format("01/06")
}

func generateCVV() string {
	return fmt.Sprintf("%03d", rand.Intn(1000))
}

func computeHMAC(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func encryptDataPGP(plaintext string, entity *openpgp.Entity) ([]byte, error) {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, openpgp.PublicKeyType, nil)
	if err != nil {
		return nil, err
	}

	encrypter, err := openpgp.Encrypt(w, []*openpgp.Entity{entity}, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	_, err = encrypter.Write([]byte(plaintext))
	if err != nil {
		return nil, err
	}

	err = encrypter.Close()
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decryptDataPGP(ciphertext []byte, entity *openpgp.Entity) (string, error) {
	block, err := armor.Decode(bytes.NewReader(ciphertext))
	if err != nil {
		return "", err
	}

	md, err := openpgp.ReadMessage(block.Body, openpgp.EntityList{entity}, nil, nil)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
