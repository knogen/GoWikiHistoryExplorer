// 匹配规则， 保存于 mode
// 这是一个中间处理程序，前置数据是已经从 wikipedia history export 提取了 reference， 并进行了去重计时处理，
// arxiv 数据从 https://www.kaggle.com/datasets/Cornell-University/arxiv 获取，直接导入 elastic ,使用默认的搜索配置，匹配模式按照打分规则如下。
// 处理预计耗时 15h
// 1：doi 匹配
// 2：title 完全匹配， 第一作者在 mag 中属于子集
// 3：title 完全匹配
// 4: title 编辑距离误差在 10% ,  作者子集匹配
// 10 title 匹配语言距离差距 5 以上, title 为空, 不保存 magID
// 99 例外情况，暂不处理
package matchArxiv

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/elastic/go-elasticsearch/v8"
	log "github.com/sirupsen/logrus"
)

func initElastic(esUrl string) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{
			esUrl,
		},
	}
	es, _ := elasticsearch.NewClient(cfg)
	return es
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Main() {

	work()
}

func work() {
	es := initElastic("http://192.168.1.226:9200")
	mongo := newMongoRefDataBase("mongodb://knogen:knogen@r730xd-2.lmd.wuzhen.ai:27017")

	linkChan := mongo.GetUnlink()
	for i := 0; i < 15; i++ {
		go func() {
			for item := range linkChan {
				log.Println("start id:", item.ID)
				// doi 匹配
				if item.Ref["doi"] != "" {

					arxiv := doiQuery(item.Ref["doi"], es)
					if arxiv != "" {
						mt := refMatch{ArxivID: arxiv, Mode: 1}
						mongo.UpdateMatch(item.ID, mt)
						continue
					}
				}
				// title 匹配
				if item.Ref["title"] != "" {

					first := item.Ref["first"]
					if first == "" {
						first = item.Ref["first1"]
					}

					last := item.Ref["last"]
					if last == "" {
						last = item.Ref["last1"]
					}
					rm := titleQuery(item.Ref["title"], first, last, es)
					if rm.Mode < 100 {
						mongo.UpdateMatch(item.ID, rm)
					}
					continue
				} else {
					// title 为空
					mt := refMatch{Mode: 10}
					mongo.UpdateMatch(item.ID, mt)
					continue
				}
			}
		}()
	}

	<-time.After(120 * time.Hour)

}

// 使用 title query elastic， 对结果按照规则进行打分，取分数最高的一个
func titleQuery(title string, first, last string, es *elasticsearch.Client) refMatch {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"size": 10,
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": title,
			},
		},
		"_source": []string{"title", "authors"},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("arxiv"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	var r map[string]interface{}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	var cacheRms []refMatch
	minDistance := 1000
	// 10个结果，按分数从高到底进行比较
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		// log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		retTitle := hit.(map[string]interface{})["_source"].(map[string]interface{})["title"].(string)

		// retYear := int(hit.(map[string]interface{})["_source"].(map[string]interface{})["year"].(float64))
		// log.Println(hit.(map[string]interface{})["_source"].(map[string]interface{}))
		authors := hit.(map[string]interface{})["_source"].(map[string]interface{})["authors"].(string)
		arxivID, _ := hit.(map[string]interface{})["_id"].(string)
		// 排除空格影响
		if strings.EqualFold(strings.ReplaceAll(retTitle, " ", ""), strings.ReplaceAll(title, " ", "")) {
			rm := refMatch{
				ArxivID: arxivID,
				Mode:    100,
			}
			if authorCheck(authors, first, last) {
				// 全部符合
				rm.Mode = 2
			} else {
				// 名字不符合
				rm.Mode = 3
			}

			cacheRms = append(cacheRms, rm)
		} else {
			// 名字不匹配
			// 进行编辑距离计算,排除空格影响，保留最小编辑距离
			distance := levenshtein.ComputeDistance(strings.ReplaceAll(strings.ToLower(retTitle), " ", ""), strings.ReplaceAll(strings.ToLower(title), " ", ""))
			if len(retTitle)/10 >= distance {
				// 编辑距离差在 10% 之内的，
				if authorCheck(authors, first, last) {
					// 全部符合
					cacheRms = append(cacheRms, refMatch{
						ArxivID: arxivID,
						Mode:    4,
					})
					continue
				}

			}
			if minDistance > distance {
				minDistance = distance
			}
		}
	}

	// 找到最优的解
	if len(cacheRms) > 0 {
		sort.SliceStable(cacheRms, func(i, j int) bool {
			return cacheRms[i].Mode < cacheRms[j].Mode
		})
		if cacheRms[0].ArxivID != "" && cacheRms[0].Mode <= 10 {
			return cacheRms[0]
		}
	}

	// title 不匹配，根据编辑距离判断是否有解
	if minDistance != 1000 && minDistance > 5 {
		rm := refMatch{Mode: 10}
		return rm
	}
	// 无法判断，暂时 不匹配
	log.Println("title not match:", title, first, last, minDistance)
	return refMatch{Mode: 99}
}

// 判断名字是否包含
func authorCheck(authors string, first, last string) bool {
	if len(authors) == 0 {
		return false
	}
	if first != "" && strings.Contains(strings.ToLower(authors), strings.ToLower(first)) {
		return true
	} else if last != "" && strings.Contains(strings.ToLower(authors), strings.ToLower(last)) {
		return true
	}
	return false
}

// 使用 dio 进行 query, doi 可信度最高
func doiQuery(doi string, es *elasticsearch.Client) (arxivID string) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"size": 1,
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"doi.keyword": doi,
			},
		},
		"_source": []string{"title", "year", "doi"},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("arxiv"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	var r map[string]interface{}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

		retDoi := hit.(map[string]interface{})["_source"].(map[string]interface{})["doi"].(string)
		arxivID, _ = hit.(map[string]interface{})["_id"].(string)
		if strings.EqualFold(doi, retDoi) {
			return
		} else {
			log.Println("not match doi:", doi, retDoi)
		}
	}
	arxivID = ""
	return
}
