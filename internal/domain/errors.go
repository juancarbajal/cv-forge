package domain

import "errors"

var (
	ErrNotFound            = errors.New("recurso no encontrado")
	ErrAlreadyExists       = errors.New("el recurso ya existe")
	ErrInvalidInput        = errors.New("datos de entrada inválidos")
	ErrUnauthorized        = errors.New("no autorizado o clave inválida")
	ErrProviderNotActive   = errors.New("no hay un proveedor de IA activo")
	ErrUnsupportedProvider = errors.New("proveedor de IA no soportado")
	ErrAntiInventionFailed = errors.New("validación anti-invención fallida: información no sustentada")
	ErrNoActivePrompt      = errors.New("no hay una versión activa para este tipo de prompt")
)
