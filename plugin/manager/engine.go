package manager

import (
	zero "ZeroBot"
)

// Engine is the pre_handler, post_handler manager
type Engine *zero.Engine

var enmap = make(map[string]*zero.Engine)

func Initialize() {
	matcher := zero.Matcher{}
	matchers := matcher.GetMatcherList()
	for _, v := range matchers {
		metaDate := v.Engine.MetaData
		if metaDate != nil {
			if metaDate.Name != "" {
				if _, ok := enmap[metaDate.Name]; ok {
					continue
				}
				enmap[metaDate.Name] = v.Engine
				v.Engine.UsePreHandler(newAuthHandler(metaDate))
				v.Engine.UsePostHandler(judgeMalicious)
			}
		}
	}
}
