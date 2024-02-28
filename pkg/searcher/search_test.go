package searcher

import (
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"
)

func TestSearcher_Search(t *testing.T) {
	type fields struct {
		FS fs.FS
	}

	type args struct {
		word string
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "Ok, case 1",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
					"file2.txt": {Data: []byte("World1")},
					"file3.txt": {Data: []byte("Hello World")},
				},
			},
			args:      args{word: "World"},
			wantFiles: []string{"file1.txt", "file3.txt"},
			wantErr:   false,
		},
		{
			name: "Ok, case 2",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World            World")},
					"file2.txt": {Data: []byte("WorldWorld")},
					"file3.txt": {Data: []byte(`- Дети, - молвил он им, - я иду в горы, хочу с другими смельчаками
поохотиться на поганого пса Алибека (так звали разбойника-турка,
разорявшего последнее время весь тот край). Ждите меня десять дней, а коли
на десятый день не вернусь, закажите вы обедню за упокой моей души -
значит, убили меня. Но ежели, - прибавил тут старый Горча, приняв вид самый
строгий, - ежели (да не попустит этого бог) я вернусь поздней, ради вашего
спасения, не впускайте вы меня в дом. Ежели будет так, приказываю вам -
забудьте, что я вам был отец, и вбейте мне осиновый кол в спину, что бы я
ни говорил, что бы ни делал, - значит, я теперь проклятый вурдалак и пришел
сосать вашу кровь. 
World`)},
				},
			},
			args:      args{word: "World"},
			wantFiles: []string{"file1.txt", "file3.txt"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Searcher{
				FS: tt.fields.FS,
			}
			err := s.ConstructFileDictionary()
			gotFiles := s.Search(tt.args.word)
			sort.Strings(gotFiles)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Search() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	type fields struct {
		FS fs.FS
	}

	type args struct {
		word string
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "not found",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
					"file2.txt": {Data: []byte("World1")},
					"file3.txt": {Data: []byte("Hello World")},
				},
			},
			args:      args{word: "keyword"},
			wantFiles: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Searcher{
				FS: tt.fields.FS,
			}
			err := s.ConstructFileDictionary()
			gotFiles := s.Search(tt.args.word)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Search() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

//  создаем моковую структуру которая реализует интерфейс FS , чтобы можно было ее передать в нашу структуру поиска файла
// её метод Open возвращает ошибку чтобы можно было протестировать случай ошибки

type FSError struct{}

func (fs FSError) Open(_ string) (fs.File, error) {
	return nil, fmt.Errorf("mock error")
}

func TestError(t *testing.T) {
	type fields struct {
		FS fs.FS
	}

	type args struct {
		word string
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "file system is broken",
			fields: fields{
				FS: FSError{},
			},
			args:      args{word: "keyword"},
			wantFiles: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Searcher{
				FS: tt.fields.FS,
			}
			err := s.ConstructFileDictionary()
			gotFiles := s.Search(tt.args.word)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Search() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}
