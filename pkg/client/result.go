package client

// TODO HRESULT constants add more???
const (
	S_OK         = uint32(0x00000000)
	E_FAIL       = uint32(0x80004005)
	E_INVALIDARG = uint32(0x80070057)
)

func IsHRESULTSuccess(hresult uint32) bool {
	return hresult == S_OK
}
func IsHRESULTFailure(hresult uint32) bool {
	return hresult != S_OK
}
