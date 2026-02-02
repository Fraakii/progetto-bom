package main

import (
	"encoding/csv" // legge i file Excel salvati come testo (CSV), gestisce virgole, punti ecc
	"encoding/json" // il traduttore. trasforma i dati da formato Go a formato Web (JSON)
	"fmt"           // format. stampa i messaggi nella console, uile per capire se funziona tutto
	"io"            // Input/Output: serve per capire quando il file è finito (end of file)
	"net/http"      // Serve a creare il Server Web che ascolta richieste del browser
	"strings"       // Serve a pulire le scritte (togliere virgolette "" inutili
)

// Definiamo come è fatto un gruppo di componenti

type GruppoComponente struct {
	Chiave      string   `json:"chiave"`		// il json tra apici è un'etichetta
	Value       string   `json:"value"`			// e dice: quando mandi un dato al browser
	Footprint   string   `json:"footprint"`		// chiamalo 'chiave' (minuscolo)
	Quantita    int      `json:"quantita"`
	Designators []string `json:"designators"`	// una Slice, una lista di stringe ["R1", "R2",]
}

func main() {
	fmt.Println("Server avviato su http://localhost:8080")

// trasforma la cartella frontend in un sito web navigabile
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

// diciamo al server "se qualcuno manda dati a /api/upload, chiama la funzione "uploadBOM".
// è il routing. collega un indirizzo web (/api/upload) a una funzione Go (uploadBOM)
	http.HandleFunc("/api/upload", uploadBOM) 

// mettiamo il server in ascolto sulla porta 8080 e blocchiamo il programma e rimane in loop
// infinito ad ascoltare le richieste finche non lo chiudi
	http.ListenAndServe(":8080", nil)
}

// ------+ FUNZIONE UPLOADBOM -------

func uploadBOM(w http.ResponseWriter, r *http.Request) {

// Configurazione CORS (Permessi Browser)
	w.Header().Set("Access-Control-Allow-Origin", "*")		// accetta dati da tutti
	w.Header().Set("Access-Control-Allow-Methods", "POST")  // POST perche senza, Chrome bloccherebbe tutto per sicurezza
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {return} // se il browser chiede "posso mandare dati?", manda una richiesta di 
									  // prova per vedere se il server è vivo. Diciamo "si" e usciamo

// Ricezione file dal Browser - cerchiamo un file chiamato file_bom dentro la richiesta (r)
	file, header, err := r.FormFile("file_bom") // file_bom sta nel codice JS custom import

	if err != nil { // dice se l'errore non è vuoto (quindi c'è l'errore)
		// se c'è un errore (es. nessun file), rispondiamo "Errore 400" al browser
		http.Error(w, "Errore ricezione: "+err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close() // chiude il file per liberare memoria e lo fa alla fine della funzione (defer)

	fmt.Printf("Ricevuto file: %s\n", header.Filename)

// CONFIGURAZIONE FILE CSV
	reader := csv.NewReader(file)   // crea uno strumento lettore e gli da il file CSV in mano

// Fondamentale per file EXCEL italiani
	reader.Comma = ';'				// usa il ; come separatore (in Italia Excel esporta CSV con ;)
	reader.FieldsPerRecord = -1		// -1 significa "non bloccarti se una riga ha 5 colonne e l'altra 6"
	reader.LazyQuotes = true 		// significa "se trovi virgolette strane, prova a ignorarle invece di crashare"


// Preparazione strutture dati
	gruppi := make(map[string]*GruppoComponente)  // creiamo la memoria RAM per raggruppare componenti uguali
												  // le Chiavi saranno stringhe (es. "10k_0805")
												  // i Valori saranno Puntatori(*) alla struttura si usano
												  // perché è piu efficiente: modifica l'oggetto originale
// Preparazione Lista finale 
// È una lista vuota che conterrà i gruppi pronti
    listaFinale := make([]*GruppoComponente, 0)

// variabili contatori per le statistiche
	rowIndex := 0
	righeLette := 0

	for {
		record, err := reader.Read()
									 // Controlla se il file è finito. EOF (end of file). è un segnale che il Sistema Operativo
		if err == io.EOF { break } 	 // manda quando non ci sono piu byte da leggere.	
								     // Quindi, se l'errore dice File finito, fermati. Sennò continua.
		if err != nil {			  						// se l'errore esiste...
			fmt.Println("Errore lettura riga:", err)	// gestisce errori diversi dalla fine del file
			continue	
		}

		if rowIndex == 0 { 		// scarta la prima riga del file
			rowIndex++			// cosi Aggiorniamo il contatore e va dalla 1 in poi
			continue
		}

		if len(record) < 5 {    // controlla se la riga ha abbastanza colonne
			continue			// len misura la lunghezza della lista. se Record è ["1", "R1", "10k"], 
								// la lunghezza è 3. Se la riga ha 3 colonne e noi chiediamo la 5 il programma va
								// in crash. Con il CONTINUE salta la riga e va alla prossima.
		}


// MAPPATURA
		// 0: Id
		// 1: Riferimento (C13,C17...)
		// 2: Impronta (C_0603...)
		// 3: Quantità (18)
		// 4: Valore (0.1u)
		
	// Qui decidiamo quali colonne leggere.
	// Se il CSV cambia ordine, basta cambiare questi numeri [1], [2], [4]
	designatorRaw := record[1] 	// Colonna Riferimento (es. "R1" o "C1,C2")
	footprint     := record[2] 	// Colonna Impronta (es. "0805")
	value         := record[4] 	// Colonna Valore (es. "10k")


// LA PULIZIA: spesso i CSV mettono virgolette inutili. Se non si levano, nel sito le vedremmo
	designatorRaw = strings.Trim(designatorRaw, "\"") // trim toglie le virgolette
    footprint     = strings.Trim(footprint, "\"")
    value         = strings.Trim(value, "\"")
        
    value = strings.TrimSpace(value)				  // trimSpace toglie gli spazi vuoti invisibili
    footprint = strings.TrimSpace(footprint)		  // all'inizio o alla fine	

	righeLette++		// Incrementiamo il contatore delle righe valide lette


// GESTIONE GRUPPI MULTIPLI (Split) 
	// Esempio: una riga dice "C1,C2,C3".
	// Dobbiamo dividerla in tre pezzi separati.
	designators := strings.Split(designatorRaw, ",")


// Cicliamo su ogni singolo designator trovato (es. prima C1, poi C2...)
		for _, d := range designators {
			// Pulizia extra per il singolo designator (spazi e virgolette residue)
			designator := strings.TrimSpace(d)
			designator = strings.Trim(designator, "\"")

			// Se dopo la pulizia è vuoto, saltiamo
			if designator == "" {
				continue
			}

			// CREAZIONE CHIAVE UNICA
			// Uniamo valore e impronta per capire se sono uguali
			chiave := value + "_" + footprint

			// LOGICA DI RAGGRUPPAMENTO
			if gruppo, esiste := gruppi[chiave]; esiste {
				// CASO A: Il gruppo esiste già. Aggiungiamo solo il pezzo.
				gruppo.Quantita++
				gruppo.Designators = append(gruppo.Designators, designator)
			} else {
				// CASO B: È un nuovo gruppo. Lo creiamo da zero.
				gruppi[chiave] = &GruppoComponente{
					Chiave:      chiave,
					Value:       value,
					Footprint:   footprint,
					Quantita:    1,
					Designators: []string{designator},
				}
			}
		}
	} 

	// PREPARAZIONE RISPOSTA (Output)
	
	// Travasiamo i dati dalla Mappa (disordinata) alla Lista (ordinata per JSON)
	for _, g := range gruppi {
		listaFinale = append(listaFinale, g)
	}

	// Log finale per noi sviluppatori
	fmt.Printf("SUCCESSO: Lette %d righe dati. Trovati %d gruppi unici.\n", righeLette, len(listaFinale))

	// Invio al Browser
	w.Header().Set("Content-Type", "application/json") // Diciamo: "Ti sto mandando un JSON"
	json.NewEncoder(w).Encode(listaFinale)             // Spediamo i dati
}



