package assets

import _ "embed"

// SuccessSound Breviceps (https://freesound.org/people/Breviceps/sounds/445978/)
// Licence: CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
//
//go:embed sounds/success.wav
var SuccessSound []byte

// FailSound PaulMorek (https://freesound.org/people/PaulMorek/sounds/330046/)
// Licence: CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
//
//go:embed sounds/fail.wav
var FailSound []byte
