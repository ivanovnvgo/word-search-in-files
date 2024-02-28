package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/ivanovnvgo/word-search-in-files/pkg/searcher"
	"go.uber.org/zap"
)

// SearchHandler structure
type SearchHandler struct {
	wordSearcher searcher.WordSearcher
	logger       *zap.SugaredLogger
}

// NewSearcherHandler creates a new object of type *SearchHandler
func NewSearcherHandler(wordSearcher searcher.WordSearcher, logger *zap.SugaredLogger) *SearchHandler {
	return &SearchHandler{
		wordSearcher: wordSearcher,
		logger:       logger,
	}
}

// Search method returns a list of file names in JSON format that contain this word
func (sh *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	// получаем ключевое слово по query параметру
	keyword := r.URL.Query().Get("keyword")
	// в случае отсутствия query параметра keyword или его пустого значения выдаем пользователю бэд реквест
	if keyword == "" {
		sh.logger.Errorf("keyword quety param is empty")
		err := writeResponse(w, http.StatusBadRequest, []byte(`{"message":"no keyword found in request"}`))
		if err != nil {
			sh.logger.Errorf("error in writing response: %s", err)
		}
		return
	}

	result := sh.wordSearcher.Search(keyword)
	// если подходящих файлов не найдено возвращаем нот фаунд
	if result == nil {
		sh.logger.Errorf("result is empty")
		err := writeResponse(w, http.StatusNotFound, []byte(`{"message":"keyword was not found in files"}`))
		if err != nil {
			sh.logger.Errorf("error in writing response: %s", err)
		}
		return
	}

	// делаем маршалинг ответа
	resultJSON, err := json.Marshal(result)
	// в случае неудачи возвращаем 500ку
	if err != nil {
		sh.logger.Errorf("error in JSON coding of result: %s", err)
		err = writeResponse(w, http.StatusInternalServerError, []byte(`{"message":"internal error"}`))
		if err != nil {
			sh.logger.Errorf("error in writing response: %s", err)
		}
		return
	}

	// формируем успешный ответ с результатом поиска
	err = writeResponse(w, http.StatusOK, resultJSON)
	if err != nil {
		sh.logger.Errorf("error in writing response: %s", err)
	}
	return
}

func writeResponse(w http.ResponseWriter, status int, respBody []byte) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_, err := w.Write(respBody)
	return err
}
