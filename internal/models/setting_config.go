package models

// SettingItem 定义设置项的配置
type SettingItem struct {
	Key          string      // 设置项的键名
	Type         string      // 数据类型：int, int64, bool, string, json
	DefaultValue interface{} // 默认值
	MethodSuffix string      // 方法名后缀，如 MultipleStreamThreadCount
}

// SettingItems 所有设置项的配置
var SettingItems = []SettingItem{
	{
		Key:          "multiple_stream_thread_count",
		Type:         "int",
		DefaultValue: 6,
		MethodSuffix: "MultipleStreamThreadCount",
	},
	{
		Key:          "multiple_stream_chunk_size",
		Type:         "int64",
		DefaultValue: int64(1024 * 1024 * 4),
		MethodSuffix: "MultipleStreamChunkSize",
	},
	{
		Key:          "strm_file_enable",
		Type:         "bool",
		DefaultValue: false,
		MethodSuffix: "StrmFileEnable",
	},
	{
		Key:          "strm_support_file_ext_list",
		Type:         "json",
		DefaultValue: []string{
			"mp4", "mkv", "avi", "mov", "wmv", "flv", "webm", "m4v",
			"mpg", "mpeg", "m2v", "m4p", "m4b", "ts", "mts", "m2ts", "m2t",
			"mxf", "dv", "dvr-ms", "asf", "3gp", "3g2", "f4v", "f4p", "f4a", "f4b",
			"vob", "ogv", "ogg", "divx", "xvid", "rm", "rmvb", "dat", "nsv",
			"qt", "amv", "mpv", "m1v", "svi", "viv", "fli", "flc",
		},
		MethodSuffix: "StrmSupportFileExtList",
	},
	{
		Key:          "link_file_auto_delete",
		Type:         "bool",
		DefaultValue: true,
		MethodSuffix: "LinkFileAutoDelete",
	},
	{
		Key:          "strm_base_url",
		Type:         "string",
		DefaultValue: "",
		MethodSuffix: "StrmBaseURL",
	},
}