package protocolo

// Heartbeat es el mensaje enviado periodicamente por UDP.
type Heartbeat struct {
	NodoID    string `json:"nodo_id"`
	Timestamp int64  `json:"timestamp"`
	Contador  int    `json:"contador"`
}

// TODO: Definir struct Lectura para RPC RegistrarLectura.
type Lectura struct {
	SensorID    string  `json:"sensor_id"`
	Temperatura float64 `json:"temperatura"`
	Timestamp   int64   `json:"timestamp"`
}

// TODO: Definir struct RespuestaLectura para respuesta de RegistrarLectura.
type RespuestaLectura struct {
	ID      int64  `json:"id"`
	Mensaje string `json:"mensaje"`
}

// TODO: Definir struct ConsultaUltimaLectura para RPC ObtenerUltimaLectura.
type ConsultaUltimaLectura struct {
	SensorID string `json:"sensor_id"`
}
