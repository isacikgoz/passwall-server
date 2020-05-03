package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pass-wall/passwall-server/internal/app"
	"github.com/pass-wall/passwall-server/internal/common"
	"github.com/pass-wall/passwall-server/internal/encryption"
	"github.com/pass-wall/passwall-server/internal/storage"
	"github.com/pass-wall/passwall-server/model"
	"github.com/spf13/viper"
)

// FindAll ...
func FindAllCreditCards(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		creditCards := []model.CreditCard{}

		fields := []string{"id", "created_at", "updated_at", "bank_name", "bank_code", "account_name", "account_number", "iban", "currency"}
		argsStr, argsInt := SetArgs(r, fields)

		creditCards, err = s.CreditCards().FindAll(argsStr, argsInt)

		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		creditCards = app.DecryptCreditCardVerificationNumbers(creditCards)
		common.RespondWithJSON(w, http.StatusOK, creditCards)
	}
}

// FindByID ...
func FindCreditCardByID(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		creditCard, err := s.CreditCards().FindByID(uint(id))
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		passByte, _ := base64.StdEncoding.DecodeString(creditCard.VerificationNumber)
		creditCard.VerificationNumber = string(encryption.Decrypt(string(passByte[:]), viper.GetString("server.passphrase")))

		common.RespondWithJSON(w, http.StatusOK, model.ToCreditCardDTO(creditCard))
	}
}

// Create ...
func CreateCreditCard(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creditCardDTO model.CreditCardDTO

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&creditCardDTO); err != nil {
			common.RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
			return
		}
		defer r.Body.Close()

		rawPass := creditCardDTO.VerificationNumber
		creditCardDTO.VerificationNumber = base64.StdEncoding.EncodeToString(encryption.Encrypt(creditCardDTO.VerificationNumber, viper.GetString("server.passphrase")))

		createdCreditCard, err := s.CreditCards().Save(model.ToCreditCard(creditCardDTO))
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		createdCreditCard.VerificationNumber = rawPass

		common.RespondWithJSON(w, http.StatusOK, model.ToCreditCardDTO(createdCreditCard))
	}
}

// Update ...
func UpdateCreditCard(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var creditCardDTO model.CreditCardDTO
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&creditCardDTO); err != nil {
			common.RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
			return
		}
		defer r.Body.Close()

		creditCard, err := s.CreditCards().FindByID(uint(id))
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		rawPass := creditCardDTO.VerificationNumber
		creditCardDTO.VerificationNumber = base64.StdEncoding.EncodeToString(encryption.Encrypt(creditCardDTO.VerificationNumber, viper.GetString("server.passphrase")))

		creditCardDTO.ID = uint(id)
		creditCard = model.ToCreditCard(creditCardDTO)
		creditCard.ID = uint(id)

		updatedCreditCard, err := s.CreditCards().Save(creditCard)
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		updatedCreditCard.VerificationNumber = rawPass
		common.RespondWithJSON(w, http.StatusOK, model.ToCreditCardDTO(updatedCreditCard))
	}
}

// Delete ...
func DeleteCreditCard(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		creditCard, err := s.CreditCards().FindByID(uint(id))
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		err = s.CreditCards().Delete(creditCard.ID)
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		response := model.Response{http.StatusOK, "Success", "CreditCard deleted successfully!"}
		common.RespondWithJSON(w, http.StatusOK, response)
	}
}
