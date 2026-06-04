package nodo

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Configuracion representa los parámetros del nodo IoT.
type Configuracion struct {
	ID                string
	Edificio          string
	Aula              string
	BrokerMQTT        string
	BrokerURL         string
	IntervaloSegundos time.Duration
}

// CargarConfiguracion lee variables de entorno o usa valores por defecto.
func CargarConfiguracion() Configuracion {
	id := obtenerEnv("NODO_ID", "nodo-01")
	edificio := obtenerEnv("NODO_EDIFICIO", "ingenieria")
	aula := obtenerEnv("NODO_AULA", "lab3")
	broker := obtenerEnv("MQTT_BROKER", "localhost:1884")
	intervalo := obtenerEnv("INTERVALO_SEGUNDOS", "5")

	// --- VALIDACIONES OBLIGATORIAS ---

	regexValido := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	if id == "" || !regexValido.MatchString(id) {
		log.Fatalf("[Config-Error] El NODO_ID '%s' es invalido. No debe estar vacio y solo puede contener letras, numeros y guiones.", id)
	}
	if edificio == "" || !regexValido.MatchString(edificio) {
		log.Fatalf("[Config-Error] El NODO_EDIFICIO '%s' es invalido. No debe estar vacio y solo puede contener letras, numeros y guiones.", edificio)
	}
	if aula == "" || !regexValido.MatchString(aula) {
		log.Fatalf("[Config-Error] El NODO_AULA '%s' es invalido. No debe estar vacio y solo puede contener letras, numeros y guiones.", aula)
	}

	intIntervalo, err := strconv.Atoi(intervalo)
	if err != nil || intIntervalo <= 0 {
		log.Fatalf("[Config-Error] El INTERVALO_SEGUNDOS '%s' es inválido. Debe ser un número entero estrictamente positivo.", intervalo)
	}

	duracion := time.Duration(intIntervalo) * time.Second

	brokerURL := broker
	if !strings.HasPrefix(brokerURL, "tcp://") && !strings.HasPrefix(brokerURL, "ssl://") && !strings.HasPrefix(brokerURL, "ws://") {
		brokerURL = fmt.Sprintf("tcp://%s", brokerURL)
	}

	return Configuracion{
		ID:                id,
		Edificio:          edificio,
		Aula:              aula,
		BrokerMQTT:        broker,
		BrokerURL:         brokerURL,
		IntervaloSegundos: duracion,
	}
}

func obtenerEnv(clave, valorPorDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return valorPorDefecto
}
