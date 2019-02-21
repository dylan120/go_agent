package utils

var (
	EXNouser                = 67 // addressee unknown
	Success                 = 0
	Failure                 = 1
	KeepAlive               = 99
	Running                 = 1001 // 正在执行 || 命令成功发送到minion客户端 || 仍有文件需要下载
	Wait                    = 1002
	Stop                    = 1003
	Ignore                  = 1004
	Killed                  = 1005
	Timeout                 = 2003
	MinionDown              = 2013
	CheckAliveDetectNoStart = 2015
	SourceFileDoNotExist    = 2006
	WrongMD5                = 2011
	InvalidVersion          = 2018
	FileDownloadFailure     = 2101
	FileMD5Error            = 2102
	TorrentSSLKeyNoExist    = 2103
)
