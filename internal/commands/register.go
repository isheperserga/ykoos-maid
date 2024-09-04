package commands

type RegisterCommand func()

var registrations []RegisterCommand

func AddRegistration(reg RegisterCommand) {
	registrations = append(registrations, reg)
}

func RegisterAll() {
	for _, reg := range registrations {
		reg()
	}
}
