package hilo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ravendb/ravendb-go-client/data"
	executor "github.com/ravendb/ravendb-go-client/http"
	"github.com/ravendb/ravendb-go-client/http/commands"
	SrvNodes "github.com/ravendb/ravendb-go-client/http/server_nodes"
	"github.com/ravendb/ravendb-go-client/tools/types"
)

/// RangeValue. The result of a NextHiLo operation
type RangeValue struct {
	min_id, max_id, current uint
}

// NewRangeValue default values min_id = 1, max_id = 0
func NewRangeValue(min_id, max_id uint) *RangeValue {

	//todo: для этой структуры я решил, что лучше просто выставлять умолчаемые значения, чем проверять ошибку - она же вспомогательная
	if min_id < 1 {
		min_id = 1
	}
	return &RangeValue{min_id, max_id, min_id - 1}
}
func RangeValueDefault() *RangeValue {
	return NewRangeValue(1, 0)
}

type HiLoReturnCommand struct {
	commands.RavenCommand
	tag       string
	last, end uint
}

//todo: еще нужно опреелиться с границами числовыз параметров
func NewHiLoReturnCommand(tag string, last, end uint) (*HiLoReturnCommand, error) {

	if (last < 0) || (end < 0) || (tag == "") {
		return nil, errors.New("ArgumentOutOfRangeException")
	}
	ref := &HiLoReturnCommand{}
	ref.Method = "PUT" // super(HiLoReturnCommand, self).__init__(method="PUT")
	ref.tag = tag
	ref.last = last
	ref.end = end

	return ref, nil
}
func (ref *HiLoReturnCommand) CreateRequest(sn SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf(`%s/databases/%s/hilo/return?tag=%s&end=%s&last=%s`, sn.GetUrl(), sn.GetDatabase(),
		ref.tag, ref.end, ref.last)
}

type NextHiLoCommand struct {
	commands.RavenCommand
	tag, serverTag, identityPartsSeparator string
	lastBatchSize, lastRangeMax            uint
	lastRangeAt                            time.Time
}

func NewNextHiLoCommand(tag string, lastBatchSize uint, lastRangeAt time.Time, identityPartsSeparator string, lastRangeMax uint) (*NextHiLoCommand, error) {
	if (identityPartsSeparator == "") || (tag == "") {
		return nil, errors.New("ArgumentOutOfRangeException")
	}

	ref := &NextHiLoCommand{
		tag:                    tag,
		lastBatchSize:          lastBatchSize,
		lastRangeAt:            lastRangeAt,
		identityPartsSeparator: identityPartsSeparator,
		lastRangeMax:           lastRangeMax,
	}
	ref.Method = "GET"
	return ref, nil
}
func (ref *NextHiLoCommand) CreateRequest(sn SrvNodes.IServerNode) {
	path := fmt.Sprintf(`hilo/next?tag=%s&lastBatchSize=%d&lastRangeAt=%s&identityPartsSeparator=%s&lastMax=%d`,
		ref.tag, ref.lastBatchSize, ref.lastRangeAt, ref.identityPartsSeparator, ref.lastRangeMax)
	ref.Url = fmt.Sprintf(`%s/databases/%s/%s`, sn.GetUrl(), sn.GetDatabase(), path)
}

//{"prefix": response["Prefix"], "serverTag": response["ServerTag"], "low": response["Low"],
//    "high": response["High"],
//    "last_size": response["LastSize"],
//    "last_range_at": response["LastRangeAt"]}, err
type nextHiloCommandJSON struct {
	prefix      string `json:Prefix`
	serverTag   string `json:ServerTag`
	low         int    `json:Low`
	high        int    `json:High`
	lastSize    int    `json:LastSize`
	LastRangeAt int    `json:LastRangeAt`
}

func (ref *NextHiLoCommand) GetResponseRaw(response *http.Response) (out []byte, err error) {
	if response.StatusCode == 201 {
		buf, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		//todo: по идее здесь надо сперва парсиь респонс, а потом маршалить, но я пока не нашел - где описано, как готовится ответ сервеа
		return json.Marshal(buf)
	}
	if response.StatusCode == 500 {
		panic(errors.New(`exceptions.DatabaseDoesNotExistException(response.json(`))
	}
	if response.StatusCode == 409 {
		panic(errors.New(`exceptions.FetchConcurrencyException(response.json(`))
	}
	return nil, errors.New(`exceptions.ErrorResponseException("Something is wrong with the request"`)
}

type MultiDatabaseHiLoKeyGenerator struct {
	DefaultDBName, url string
	conventions        *data.DocumentConvention
	generators         map[string]*MultiTypeHiLoKeyGenerator
}

func NewMultiDatabaseHiLoKeyGenerator(dbName, url string, conventions *data.DocumentConvention) *MultiDatabaseHiLoKeyGenerator {
	return &MultiDatabaseHiLoKeyGenerator{dbName, url, conventions,
		make(map[string]*MultiTypeHiLoKeyGenerator, 0),
	}
}
func (ref *MultiDatabaseHiLoKeyGenerator) GenerateDocumentKey(dbName string, entity types.TDocByEntity) string {
	if dbName == "" {
		dbName = ref.DefaultDBName
	}
	generator, ok := ref.generators[dbName]
	if !ok {
		generator = NewMultiTypeHiLoKeyGenerator(*ref)
		ref.generators[dbName] = generator
	}

	return generator.generateDocumentKey(entity)
}
func (ref *MultiDatabaseHiLoKeyGenerator) returnUnusedRange() {
	for key := range ref.generators {
		ref.generators[key].returnUnusedRange()
	}
}
func (obj MultiDatabaseHiLoKeyGenerator) createExecutor() (*executor.RequestExecutor, error) {
	return executor.CreateForSingleNode(obj.url, obj.DefaultDBName)
}

// MultiTypeHiLoKeyGenerator Generate hilo numbers against a RavenDB document
type MultiTypeHiLoKeyGenerator struct {
	parent              MultiDatabaseHiLoKeyGenerator
	keyGeneratorsByTags map[string]*HiLoKeyGenerator
	lock                sync.RWMutex
}

func NewMultiTypeHiLoKeyGenerator(parent MultiDatabaseHiLoKeyGenerator) *MultiTypeHiLoKeyGenerator {
	return &MultiTypeHiLoKeyGenerator{
		parent:              parent,
		keyGeneratorsByTags: make(map[string]*HiLoKeyGenerator, 0),
	}
}

/// Generates the document ID.
//todo: пока не понимаю -как формируется этот таг
func (ref *MultiTypeHiLoKeyGenerator) generateDocumentKey(entity types.TDocByEntity) string {
	tag := ref.parent.conventions.DefaultTransformTypeTagName(entity.Key) //.class.name)
	if tag == "" {
		return ""
	}
	value, ok := ref.keyGeneratorsByTags[tag]
	if !ok {
		// with ref.lock:
		if value, ok = ref.keyGeneratorsByTags[tag]; !ok {
			value = NewHiLoKeyGenerator(tag, *ref)
			ref.lock.Lock()
			defer ref.lock.Unlock()
			ref.keyGeneratorsByTags[tag] = value
		}
	}
	return value.generateDocumentKey()
}
func (ref *MultiTypeHiLoKeyGenerator) returnUnusedRange() {
	for key := range ref.keyGeneratorsByTags {
		ref.keyGeneratorsByTags[key].returnUnusedRange()
	}
}

type HiLoKeyGenerator struct {
	tag, prefix, server_tag string
	last_batch_size         uint
	parent                  MultiTypeHiLoKeyGenerator
	rangeValues             *RangeValue
	collection_ranges       map[string]*RangeValue
	last_range_at           time.Time
	lock                    sync.RWMutex
}

func NewHiLoKeyGenerator(tag string, parent MultiTypeHiLoKeyGenerator) *HiLoKeyGenerator {
	ref := &HiLoKeyGenerator{}
	ref.tag = tag
	ref.parent = parent
	//ref.last_range_at = datetime(1, 1, 1)
	ref.rangeValues = NewRangeValue(1, 0)
	ref.collection_ranges = make(map[string]*RangeValue, 0)
	return ref
}
func (ref *HiLoKeyGenerator) getDocumentKeyFromId(next_id uint) string {
	return fmt.Sprintf(`%s%s-%d`, ref.prefix, next_id, ref.server_tag)
}

/// Generates the document ID.
func (ref *HiLoKeyGenerator) generateDocumentKey() string {
	// lock this until not change range currently
	ref.lock.Lock()
	defer ref.lock.Unlock()
	ref.collection_ranges[ref.tag] = RangeValueDefault()
	ref.next_id()
	return ref.getDocumentKeyFromId(ref.collection_ranges[ref.tag].current)
}
func (ref *HiLoKeyGenerator) next_id() {
	// todo: убей не пойму - зачем здесь цикл , если мы блокируем все выше
	//for  true {
	my_collection_range, ok := ref.collection_ranges[ref.tag]
	if !ok {
		return
	}
	my_collection_range.current++
	if my_collection_range.current <= my_collection_range.max_id {
		return
	}
	//// with ref.lock:
	//if my_collection_range != ref.collection_ranges[ref.tag] {
	//    continue
	//}
	ref.getNextRange()
	ref.collection_ranges[ref.tag] = ref.rangeValues
	// exceptions.FetchConcurrencyException
	//}
	//}

}

//todo: не стал описывать парсинг структуры, возможно, это вообще лишнее и стоит заменить на просто чтение прямо из тела ответа??
type resHiloCommand struct {
	prefix, serverTag, last_range_at string
	low, high, last_size             uint
}

func (ref *HiLoKeyGenerator) getNextRange() error {
	hilo_command, err := NewNextHiLoCommand(ref.tag, ref.last_batch_size, ref.last_range_at, ref.parent.parent.conventions.IdentityPartsSeparator, ref.rangeValues.max_id)
	if err != nil {
		return err
	}
	executor, err := ref.parent.parent.createExecutor()
	if err == nil {
		buf, err := executor.ExecuteOnCurrentNode(hilo_command, true)
		if err != nil {
			return err
		}
		var result resHiloCommand
		err = json.Unmarshal(buf, result)
		if err != nil {
			return err
		}
		ref.prefix = result.prefix
		ref.server_tag = result.serverTag
		//todoL возможно, ято здесь сервер будет сразу отдавать в нужном формате и ненужно преоборазование?
		ref.last_range_at, err = time.Parse("2006-01-02 15:04:05 -0700 MST", result.last_range_at)
		if err != nil {
			return err
		}
		ref.last_batch_size = result.last_size
		ref.rangeValues = NewRangeValue(result.low, result.high)
	}

	return err
}
func (ref *HiLoKeyGenerator) returnUnusedRange() error {
	return_command, err := NewHiLoReturnCommand(ref.tag, ref.rangeValues.current, ref.rangeValues.max_id)
	if err != nil {
		return err
	}
	//ref.store.GetRequestExecutor("")
	executor, err := ref.parent.parent.createExecutor()
	if err == nil {
		// todo: мы никак не используем и не проверяем возврат байтов?
		_, err = executor.ExecuteOnCurrentNode(return_command, true)
	}

	return err

}
