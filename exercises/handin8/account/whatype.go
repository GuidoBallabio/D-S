package account

// WhatType discerns what type a struct is
type WhatType interface {
	WhatType() string
}
