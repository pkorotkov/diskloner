package definitions

// var Permission struct {
// 	O_R    os.FileMode
// 	OG_R   os.FileMode
// 	A_R    os.FileMode
// 	O_RW   os.FileMode
// 	OG_RW  os.FileMode
// 	A_RW   os.FileMode
// 	O_RWX  os.FileMode
// 	OG_RWX os.FileMode
// 	A_RWX  os.FileMode
// }

var AppPath struct {
	Progress string
	Log      string
}

func init() {
	// Permission.O_R = os.FileMode(0400)
	// Permission.OG_R = os.FileMode(0440)
	// Permission.A_R = os.FileMode(0444)
	// Permission.O_RW = os.FileMode(0600)
	// Permission.OG_RW = os.FileMode(0660)
	// Permission.A_RW = os.FileMode(0666)
	// Permission.O_RWX = os.FileMode(0700)
	// Permission.OG_RWX = os.FileMode(0770)
	// Permission.A_RWX = os.FileMode(0777)

	AppPath.Progress = "/var/lib/diskloner/progress"
	AppPath.Log = "/var/log/diskloner/log"
}
