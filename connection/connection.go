package connection

type Connection struct {
	Address      string `yaml:"address,omitempty"`
	User         string `yaml:"user,omitempty"`
	Description  string `yaml:"comment,omitempty"`
	IdentityFile string `yaml:"identity,omitempty"`
	PassFile     string `yaml:"passfile,omitempty"`
	SshProgram   string `yaml:"sshprogram,omitempty"`
}

type Item struct {
	Name string
	Conn Connection
}

func (i Item) FinalAddr() string {
	finalAddr := ""

	if i.Conn.User != "" {
		finalAddr = i.Conn.User + "@"
	}
	if i.Conn.Address != "" {
		finalAddr += i.Conn.Address
	} else {
		finalAddr += i.Name
	}

	return finalAddr
}

func (i Item) FilterValue() string {
	return i.Name + " " + i.Conn.Address + " " + i.Conn.User + " " + i.Conn.Description
}

func (i Item) Title() string {
	return i.Name
}

func (i Item) Description() string {
	return i.FinalAddr() + " -> " + i.Conn.Description
}
