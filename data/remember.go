package data

type Remember interface {
	PrepareNext(*Card, int8)
}
