package httputils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"erguotou520/sub2singbox/convert"
	"erguotou520/sub2singbox/model/clash"

	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

func GetClash(ctx context.Context, hc *http.Client, urls []string, addTag bool, include, exclude string) (clash.Clash, error) {
	c, _, _, err := GetAny(ctx, hc, urls, addTag, include, exclude)
	if err != nil {
		return c, fmt.Errorf("GetClash: %w", err)
	}
	return c, nil
}

func GetAny(ctx context.Context, hc *http.Client, urls []string, addTag bool, include, exclude string) (clash.Clash, []map[string]any, []string, error) {
	c := clash.Clash{}
	singList := []map[string]any{}
	tags := []string{}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(3)

	l := sync.Mutex{}

	for _, v := range urls {
		v := v
		u, err := url.Parse(v)
		if err != nil {
			return c, nil, nil, fmt.Errorf("GetAny: %w", err)
		}
		host := u.Host
		g.Go(func() error {
			b, err := HttpGet(ctx, hc, v, 1000*1000*10)
			if err != nil {
				return err
			}
			lc := clash.Clash{}
			err = yaml.Unmarshal(b, &lc)
			if err != nil || len(lc.Proxies) == 0 {
				h := ""
				if addTag {
					h = host
				}
				s, t, err := getSing(b, h, include, exclude)
				if err != nil {
					return err
				}
				l.Lock()
				singList = append(singList, s...)
				tags = append(tags, t...)
				l.Unlock()
			}
			if addTag {
				lc.Proxies = lo.Map(lc.Proxies, func(item clash.Proxies, index int) clash.Proxies {
					item.Name = fmt.Sprintf("%s[%s]", item.Name, host)
					return item
				})
				lc.ProxyGroup = lo.Map(lc.ProxyGroup, func(item clash.ProxyGroup, index int) clash.ProxyGroup {
					item.Proxies = lo.Map(item.Proxies, func(item string, index int) string {
						return fmt.Sprintf("%s[%s]", item, host)
					})
					return item
				})
			}
			l.Lock()
			defer l.Unlock()
			c.Proxies = append(c.Proxies, lc.Proxies...)
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return c, nil, nil, fmt.Errorf("GetAny: %w", err)
	}
	return c, singList, tags, nil
}

var ErrJson = errors.New("错误的格式")

func getSing(config []byte, host string, include, exclude string) ([]map[string]any, []string, error) {
	if !gjson.Valid(string(config)) {
		return nil, nil, fmt.Errorf("getSing: %w", ErrJson)
	}

	out := gjson.GetBytes(config, "outbounds").Array()
	outList := make([]map[string]any, 0, len(out))
	tagsList := make([]string, 0, len(out))

	for _, v := range out {
		outtype := v.Get("type").String()
		if _, ok := notNeedType[outtype]; ok {
			continue
		}
		m, ok := v.Value().(map[string]any)
		if ok {
			tag := v.Get("tag").String()
			if include != "" {
				leftTags, err := convert.Filter(true, include, []string{tag})
				if err != nil || len(leftTags) == 0 {
					continue
				}
			}
			if exclude != "" {
				leftTags, err := convert.Filter(false, exclude, []string{tag})
				if err != nil || len(leftTags) == 0 {
					continue
				}
			}
			if host != "" {
				tag = fmt.Sprintf("%s[%s]", tag, host)
				m["tag"] = tag
			}
			outList = append(outList, m)
			if outtype != "shadowtls" {
				tagsList = append(tagsList, tag)
			}
		}
	}
	return outList, tagsList, nil
}

var notNeedType = map[string]struct{}{
	"direct":   {},
	"block":    {},
	"dns":      {},
	"selector": {},
	"urltest":  {},
}
