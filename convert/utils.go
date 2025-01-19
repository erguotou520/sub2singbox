package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"erguotou520/sub2singbox/model/clash"
	"erguotou520/sub2singbox/model/singbox"

	"github.com/samber/lo"
)

func Filter(isinclude bool, reg string, sl []string) ([]string, error) {
	r, err := regexp.Compile(reg)
	if err != nil {
		return sl, fmt.Errorf("filter: %w", err)
	}
	return getForList(sl, func(v string) (string, bool) {
		has := r.MatchString(v)
		if has && isinclude {
			return v, true
		}
		if !isinclude && !has {
			return v, true
		}
		return "", false
	}), nil
}

func getForList[K, V any](l []K, check func(K) (V, bool)) []V {
	sl := make([]V, 0, len(l))
	for _, v := range l {
		s, ok := check(v)
		if !ok {
			continue
		}
		sl = append(sl, s)
	}
	return sl
}

// func getServers(s []singbox.SingBoxOut) []string {
// 	m := map[string]struct{}{}
// 	return getForList(s, func(v singbox.SingBoxOut) (string, bool) {
// 		server := v.Server
// 		_, has := m[server]
// 		if server == "" || has {
// 			return "", false
// 		}
// 		m[server] = struct{}{}
// 		return server, true
// 	})
// }

func getTags(s []singbox.SingBoxOut) []string {
	return getForList(s, func(v singbox.SingBoxOut) (string, bool) {
		tag := v.Tag
		if tag == "" || v.Ignored || len(v.Visible) != 0 {
			return "", false
		}
		return tag, true
	})
}

func Patch(b []byte, s []singbox.SingBoxOut, include, exclude string, extOut []interface{}, extags ...string) ([]byte, error) {
	d, err := PatchMap(b, s, include, exclude, extOut, extags, true, true)
	if err != nil {
		return nil, fmt.Errorf("Patch: %w", err)
	}
	bw := &bytes.Buffer{}
	jw := json.NewEncoder(bw)
	jw.SetIndent("", "    ")
	err = jw.Encode(d)
	if err != nil {
		return nil, fmt.Errorf("Patch: %w", err)
	}
	return bw.Bytes(), nil
}

func ToInsecure(c *clash.Clash) {
	for i := range c.Proxies {
		p := c.Proxies[i]
		p.SkipCertVerify = true
		c.Proxies[i] = p
	}
}

func PatchMap(
	tpl []byte,
	s []singbox.SingBoxOut,
	include, exclude string,
	extOut []interface{},
	extags []string,
	urltestOut bool,
	outFields bool,
) (map[string]any, error) {
	d := map[string]interface{}{}
	err := json.Unmarshal(tpl, &d)
	if err != nil {
		return nil, fmt.Errorf("PatchMap: %w", err)
	}
	tags := getTags(s)

	tags = append(tags, extags...)

	ftags := tags
	if include != "" {
		ftags, err = Filter(true, include, ftags)
		if err != nil {
			return nil, fmt.Errorf("PatchMap: %w", err)
		}
	}
	if exclude != "" {
		ftags, err = Filter(false, exclude, ftags)
		if err != nil {
			return nil, fmt.Errorf("PatchMap: %w", err)
		}
	}

	anyList := make([]any, 0, len(s)+len(extOut)+5)

	if urltestOut {
		anyList = append(anyList, singbox.SingBoxOut{
			Type:      "selector",
			Tag:       "select",
			Outbounds: append([]string{"urltest"}, tags...),
			Default:   "urltest",
		})
		anyList = append(anyList, singbox.SingBoxOut{
			Type:      "urltest",
			Tag:       "urltest",
			Outbounds: ftags,
		})
	}

	anyList = append(anyList, extOut...)
	for _, v := range s {
		anyList = append(anyList, v)
	}

	anyList = append(anyList, singbox.SingBoxOut{
		Type: "direct",
		Tag:  "direct",
	})

	if outFields {
		anyList = append(anyList, singbox.SingBoxOut{
			Type: "block",
			Tag:  "block",
		})
		anyList = append(anyList, singbox.SingBoxOut{
			Type: "dns",
			Tag:  "dns-out",
		})
	}

	d["outbounds"] = anyList

	return d, nil
}

func Template(templateSB []byte, singList []map[string]any, tags []string) ([]byte, error) {
	d := map[string]interface{}{}
	err := json.Unmarshal(templateSB, &d)
	if err != nil {
		return nil, fmt.Errorf("PatchMap: %w", err)
	}
	// 获取d 的 outbounds
	outbounds := d["outbounds"].([]any)
	for _, v := range outbounds {
		item, _ := v.(map[string]any)
		if item["outbounds"] == nil {
			continue
		}
		subOutbounds := item["outbounds"].([]any)
		index := lo.IndexOf(subOutbounds, "{all}")
		if index != -1 {
			// 从index处开始，往后面插入 all 对应的 tags，
			// 在此之前需要先根据include和exclude过滤tags
			allTags := tags[:]
			if item["filter"] != nil {
				filters := item["filter"].([]any)
				if len(filters) > 0 {
					for _, filter := range filters {
						filterItem, _ := filter.(map[string]any)
						action := filterItem["action"].(string)
						keywords := filterItem["keywords"].([]interface{})
						if action == "include" {
							// 使用临时变量存储所有匹配的tag
							matchedTags := make([]string, 0)
							for _, keyword := range keywords {
								filtered, _ := Filter(true, keyword.(string), allTags)
								matchedTags = append(matchedTags, filtered...)
							}
							// 去重
							allTags = lo.Uniq(matchedTags)
						} else if action == "exclude" {
							for _, keyword := range keywords {
								allTags, _ = Filter(false, keyword.(string), allTags)
							}
						}
					}
				}
				delete(item, "filter")
			}
			// 将 allTags 转换为 []any 类型
			anyTags := make([]any, len(allTags))
			for i, tag := range allTags {
				anyTags[i] = tag
			}
			item["outbounds"] = append(subOutbounds[:index], append(anyTags, subOutbounds[index+1:]...)...)
		}

	}

	// 插入所有的 singbox.SingBoxOut
	for _, v := range singList {
		outbounds = append(outbounds, v)
	}

	d["outbounds"] = outbounds

	return json.MarshalIndent(d, "", "  ")
}
