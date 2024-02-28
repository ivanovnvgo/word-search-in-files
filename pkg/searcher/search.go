package searcher

import (
	"bufio"
	"fmt"
	"io/fs"
	"sync"

	"github.com/ivanovnvgo/word-search-in-files/pkg/internal/dir"
)

// Searcher structure
type Searcher struct {
	FS             fs.FS
	Dir            string
	FileDictionary map[string]map[string]struct{}
}

// WordSearcher interface
type WordSearcher interface {
	Search(word string) []string
	ConstructFileDictionary() error
}

// NewSearcher creates a new object of type *Searcher
func NewSearcher(fs fs.FS, dir string) *Searcher {
	return &Searcher{
		FS:  fs,
		Dir: dir,
	}
}

// ConstructFileDictionary forms a map from maps containing the entire word in the files.
// The external map key is the file name, the internal map key is the words contained in this file
func (s *Searcher) ConstructFileDictionary() error {
	s.FileDictionary = make(map[string]map[string]struct{})
	// добавляем мьютекс для конкурентного доступа к мапе файлов при параллельном добавлении в нее
	mu := &sync.Mutex{}
	// получаем все имена файлов лежащих в данной директории
	fileNames, err := dir.FilesFS(s.FS, s.Dir)
	if err != nil {
		return fmt.Errorf("error from ConstructFileDictionary: %v", err)
	}
	// добавляем waitgroup для того чтобы основной поток ждал окончания работы всех горутин
	wg := &sync.WaitGroup{}
	// добавляем канал ошибок в который будет писать каждая горутина при вызове return
	// это для того, чтобы если в какой то из них случилась ошибка, например, при открывании файла,
	// получилось выйти из внешней функции
	errChan := make(chan error)
	// сразу добавляем вейтгруппе счетчик, равный длине слайса файлов
	wg.Add(len(fileNames))
	for _, fName := range fileNames {
		s.FileDictionary[fName] = make(map[string]struct{})
		// на каждый файл запускается горутина, позволяющая параллельно искать слово по файлам
		// в каждую горутину передаем копию fName чтобы избежать нежелаетельного неверного чтения при замыкании
		go func(fName string) {
			// ошибка, которая будет писаться в канал ошибок по завершении горутины
			var errorGoroutine error
			defer wg.Done()
			// открываем файл
			file, errorGoroutine := s.FS.Open(fName)
			// если возникла ошибка, значит происходит return и пишем значение ошибки в канал, которое != nil
			if errorGoroutine != nil {
				errChan <- errorGoroutine
				return
			}
			// в дефере закрываем файл во избежание утечки ресурсов
			defer func(file fs.File) {
				err := file.Close()
				if err != nil {
					return
				}
			}(file)
			// создаем сканнер, чтобы построчно считать файл, не считываем сразу весь,
			// так как в случае его большого размера пришлось бы использовать много памяти, чего мы избегаем
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				// инициализируем текущее слово, в моей программе текущее слово проверяется на равенство заданному,
				// когда оно закончилось
				// в моей программе разделителем слов служит пробел, запятая, точка и конец строки
				currentWord := make([]byte, 0)
				for i := 0; i < len(line); i++ {
					if line[i] == ' ' || line[i] == ',' || line[i] == '.' {
						if len(currentWord) != 0 {
							mu.Lock()
							s.FileDictionary[fName][string(currentWord)] = struct{}{}
							mu.Unlock()
							currentWord = currentWord[:0]
						}
					} else {
						currentWord = append(currentWord, line[i])
					}
				}
				if len(currentWord) != 0 {
					mu.Lock()
					s.FileDictionary[fName][string(currentWord)] = struct{}{}
					mu.Unlock()
				}
			}
			if errorGoroutine = scanner.Err(); errorGoroutine != nil {
				errChan <- errorGoroutine
				return
			}

		}(fName)
	}
	// в горутине ждем снятия блока на waitgroup и закрываем канал ошибок, чтобы завершился цикл по каналу
	// и можно было завершить функцию,
	// делаем это в горутине, чтобы не заблочиться и пройтись по каналу далее канал закрывается тогда,
	// когда из него было прочитано количество раз, равное количеству файлов, то есть нет риска записи в закрытый канал
	go func() {
		wg.Wait()
		close(errChan)
	}()
	// данный цикл по каналу нужен, чтобы после каждой завершившейся горутины по поиску слова в файле проверить
	// не было ли ошибки и в случае если была завершить функцию с ошибкой
	for err = range errChan {
		return fmt.Errorf("error opening/reading file from ConstructFileDictionary: %v", err)
	}
	return nil

}

// Search generates a list of files containing a keyword
func (s *Searcher) Search(word string) []string {
	files := make([]string, 0)
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for fName := range s.FileDictionary {
		wg.Add(1)
		go func(fName string) {
			defer wg.Done()
			if _, exists := s.FileDictionary[fName][word]; exists {
				mu.Lock()
				files = append(files, fName)
				mu.Unlock()
			}
		}(fName)
	}
	wg.Wait()
	if len(files) == 0 {
		return nil
	}
	return files
}
