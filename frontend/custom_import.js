
console.log("Custom Import Script (Tile + Fix) Caricato.");


// CONFIGURAZIONE

const CONFIG = {
    testoPulsante: "Import",
    icona: "📂",
    dimensioneIcona: "24px",
    dimensioneTesto: "11px",
    
    coloreSfondo: "#eeeeee",
    coloreTesto: "#333333",
    bordo: "1px solid #a0a0a0",       
    
    margineDestro: "10px",            
    margineSinistro: "5px",

    margineSuperiore: "5px", // Aumenta questo numero per scendere (es. "8px")
    
    posizione: "inizio"               
};

function creaPulsante() {
    var headerBar = document.getElementById("bomcontrols");
    
    if (!headerBar) {
        console.warn("Header non trovato.");
        return;
    }

    var btn = document.createElement("button");
    btn.className = "btn-import-custom"; 
    
    // --- STILE "TILE" ---
    btn.style.display = "inline-flex";       
    btn.style.flexDirection = "column";      
    btn.style.alignItems = "center";         
    btn.style.justifyContent = "center";
    btn.style.gap = "2px";                   
    
    btn.style.height = "46px";               
    btn.style.minWidth = "50px";             
    btn.style.padding = "4px 8px";           
    
    btn.style.cursor = "pointer";
    btn.style.borderRadius = "4px";
    btn.style.transition = "all 0.2s ease";  
    
    // Applicazione Configurazione
    btn.style.backgroundColor = CONFIG.coloreSfondo;
    btn.style.color = CONFIG.coloreTesto;
    btn.style.border = CONFIG.bordo;
    btn.style.marginRight = CONFIG.margineDestro;
    btn.style.marginLeft = CONFIG.margineSinistro;
    
    // --- ECCO LA RIGA CHE MANCAVA PRIMA! ---
    btn.style.marginTop = CONFIG.margineSuperiore; 

    // Icona
    var spanIcona = document.createElement("span");
    spanIcona.innerHTML = CONFIG.icona;
    spanIcona.style.fontSize = CONFIG.dimensioneIcona;
    spanIcona.style.lineHeight = "1"; 
    spanIcona.style.marginBottom = "0px";

    // Testo
    var spanTesto = document.createElement("span");
    spanTesto.innerHTML = CONFIG.testoPulsante;
    spanTesto.style.fontSize = CONFIG.dimensioneTesto;
    spanTesto.style.fontWeight = "600";
    spanTesto.style.lineHeight = "1"; 

    btn.appendChild(spanIcona);
    btn.appendChild(spanTesto);

    // Effetti Hover
    btn.onmouseover = function() { 
        this.style.backgroundColor = "#d0d0d0"; 
        this.style.borderColor = "#666";
        this.style.transform = "translateY(-1px)"; 
    };
    btn.onmouseout = function() { 
        this.style.backgroundColor = CONFIG.coloreSfondo; 
        this.style.borderColor = "#a0a0a0";
        this.style.transform = "translateY(0)";
    };

    // Input File nascosto
    var input = document.createElement("input");
    input.type = "file";
    input.accept = ".csv";
    input.style.display = "none";
    input.id = "fileInputBOM";

    // Azioni
    btn.onclick = function() { document.getElementById("fileInputBOM").click(); };
    input.addEventListener('change', gestisciFile);

    // Inserimento
    if (CONFIG.posizione === "fine") {
        headerBar.prepend(btn); 
    } else {
        headerBar.appendChild(btn); 
    }
    
    document.body.appendChild(input);
}

// LOGICA CARICAMENTO
function gestisciFile() {
    var file = this.files[0];
    if (!file) return;

    var btn = document.querySelector(".btn-import-custom");
    var spanTesto = btn.querySelectorAll("span")[1]; 
    var testoOriginale = spanTesto.innerHTML;
    
    spanTesto.innerHTML = "Wait..."; 
    btn.style.backgroundColor = "#fff3cd"; 

    var formData = new FormData();
    formData.append("file_bom", file);

    fetch('http://localhost:8080/api/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) throw new Error("Errore Server");
        return response.json();
    })
    .then(data => {
        try {
            applicaDatiAIBOM(data);
            spanTesto.innerHTML = "OK!";
            btn.style.backgroundColor = "#d4edda"; 
            btn.style.borderColor = "#28a745";
        } catch (e) {
            console.error(e);
            alert(e.message);
            spanTesto.innerHTML = "Err";
        }
        setTimeout(() => {
            spanTesto.innerHTML = testoOriginale;
            btn.style.backgroundColor = CONFIG.coloreSfondo;
            btn.style.borderColor = CONFIG.bordo;
        }, 2000);
    })
    .catch(err => {
        alert("Errore: " + err);
        spanTesto.innerHTML = "No Net";
        btn.style.backgroundColor = "#f8d7da"; 
    });
}

function applicaDatiAIBOM(datiBackend) {
    if (typeof pcbdata === 'undefined' || !pcbdata.footprints) {
        throw new Error("Errore critico: pcbdata non trovato.");
    }

    var refTofootprintIndex = {};
    pcbdata.footprints.forEach((fp, index) => {
        refTofootprintIndex[fp.ref] = index;
    });

    var valueColIndex = config.fields.indexOf("Value");
    var footprintColIndex = config.fields.indexOf("Footprint");

    var nuovaBomBoth = [];
    
    datiBackend.forEach((gruppo) => {
        var entryRefs = []; 
        gruppo.designators.forEach(ref => {
            var index = refTofootprintIndex[ref];
            if (index !== undefined) {
                entryRefs.push([ref, index]);
                if (valueColIndex >= 0) pcbdata.bom.fields[index][valueColIndex] = gruppo.value;
                if (footprintColIndex >= 0) pcbdata.bom.fields[index][footprintColIndex] = gruppo.footprint;
            }
        });
        if (entryRefs.length > 0) nuovaBomBoth.push(entryRefs);
    });

    pcbdata.bom.both = nuovaBomBoth;
    pcbdata.bom.F = nuovaBomBoth.filter(e => pcbdata.footprints[e[0][1]].layer === 'F');
    pcbdata.bom.B = nuovaBomBoth.filter(e => pcbdata.footprints[e[0][1]].layer === 'B');

    if (typeof populateBomTable === 'function') {
        filter = ""; 
        reflookup = "";
        populateBomTable();
        if (typeof redrawCanvas === 'function' && typeof allcanvas !== 'undefined') {
            redrawCanvas(allcanvas.front);
            redrawCanvas(allcanvas.back);
        }
    }
}

// AVVIO
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", creaPulsante);
} else {
    creaPulsante();
}