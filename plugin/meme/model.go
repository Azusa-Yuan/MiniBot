package meme

type EmojiInfo struct {
	Key        string     `json:"key"`
	ParamsType ParamsType `json:"params_type"`
	Keywords   []string   `json:"keywords"`
	Shortcuts  []Shortcut `json:"shortcuts"`
	Tags       []string   `json:"tags"`
}

type ParamsType struct {
	MinImages    int64         `json:"min_images"`
	MaxImages    int64         `json:"max_images"`
	MinTexts     int64         `json:"min_texts"`
	MaxTexts     int64         `json:"max_texts"`
	DefaultTexts []interface{} `json:"default_texts"`
	ArgsType     ArgsType      `json:"args_type"`
}

type Shortcut struct {
	Key       string      `json:"key"`
	Args      []string    `json:"args"`
	Humanized interface{} `json:"humanized"`
}

type ArgsType struct {
	// ArgsModel     ArgsModel      `json:"args_model"`
	ArgsExamples  []interface{}  `json:"args_examples"`
	ParserOptions []ParserOption `json:"parser_options"`
}

//	type ArgsModel struct {
//		Defs       Defs                `json:"$defs"`
//		// Properties ArgsModelProperties `json:"properties"`
//		Title      string              `json:"title"`
//		Type       string              `json:"type"`
//	}
// type Defs struct {
// 	UserInfo UserInfo `json:"UserInfo"`
// }

type UserInfo struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

type ParserOption struct {
	Names   []string    `json:"names"`
	Args    []Arg       `json:"args"`
	Dest    *string     `json:"dest"`
	Default interface{} `json:"default"`
	// Action   *Action     `json:"action"`
	HelpText string `json:"help_text"`
	Compact  bool   `json:"compact"`
}

type Arg struct {
	Name    string      `json:"name"`
	Value   string      `json:"value"`
	Default interface{} `json:"default"`
	Flags   interface{} `json:"flags"`
}
