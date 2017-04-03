package srnd

type ErrorModel struct {
	Err error
}

func (self *ErrorModel) Error() string {
	return self.Err.Error()
}

func (self *ErrorModel) HasError() bool {
	return self.Err != nil
}

type StepModel struct {
	Node *dialogNode
}

func (self *StepModel) HasNext() bool {
	return len(self.Node.children) > 0
}

func (self *StepModel) HasPrevious() bool {
	return self.Node.parent != nil
}

type BaseDialogModel struct {
	ErrorModel
	StepModel
}

type DBModel struct {
	ErrorModel
	StepModel

	username string
	host     string
	port     string
}

func (self *DBModel) Username() string {
	return self.username
}

func (self *DBModel) Host() string {
	return self.host
}

func (self *DBModel) Port() string {
	return self.port
}

type NameModel struct {
	ErrorModel
	StepModel

	name string
}

func (self *NameModel) Name() string {
	return self.name
}

type CryptoModel struct {
	ErrorModel
	StepModel

	host string
	key  string
}

func (self *CryptoModel) Host() string {
	return self.host
}

func (self *CryptoModel) Key() string {
	return self.key
}

type BinaryModel struct {
	ErrorModel
	StepModel

	convert string
	ffmpeg  string
	sox     string
}

func (self *BinaryModel) Convert() string {
	return self.convert
}

func (self *BinaryModel) FFmpeg() string {
	return self.ffmpeg
}

func (self *BinaryModel) Sox() string {
	return self.sox
}

type FrontendModel struct {
	ErrorModel
	StepModel

	name   string
	locale string
}

func (self *FrontendModel) Name() string {
	return self.name
}

func (self *FrontendModel) Locale() string {
	return self.locale
}

type APIModel struct {
	ErrorModel
	StepModel

	name   string
	secret string
}

func (self *APIModel) User() string {
	return self.name
}

func (self *APIModel) Secret() string {
	return self.secret
}

type KeyModel struct {
	ErrorModel
	StepModel

	public string
	secret string
}

func (self *KeyModel) Public() string {
	return self.public
}

func (self *KeyModel) Secret() string {
	return self.secret
}
