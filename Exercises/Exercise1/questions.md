Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Parallelism betyr at flere beregninger urføres fysisk samtidig, for eksempel ved at flere threads kjører på forskjellige kjerner i samme prosessor samtidig. Dette er en maskinvare-egenskap.

Concurrency handler om hvordan et program er strukturert for å håndtere flere oppgaver som er logisk overlappende i tid, selv om de ikke nødvendigvis kjører samtidig. Concurrency innebærer koordinering, synkronisering oog håndtering av delte ressurser og race conditions. Dette er en programvaredesign egenskap.

Man designer for concurrency for å oppnå parallelisme. Så parallelisme er kun den fysiske effekten, mens concurrency er designmetoden som tillater/fører til denne effekten.

What is the difference between a *race condition* and a *data race*? 
> Race condition er når et programs korrekthet avhenger av den relative timingen eller rekkefølgen av samtidige operasjoner. Programmet kan produsere ulike resultater avhengig av rekkefølgen operasjonene utføres i. 

Data race er en spesifikk type race condition der to eller flere threads aksesserer samme minneområde samtidig, minst én av operasjonene er en skriveoperasjon, og det finnes ikke tilstrekkelig synkronisering. 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> En scheduler bestemmer hvilken thread eller prosess som skal kjøre til enhver tid. Den allokerer cpu tid ved å switche mellom kjørbare threads i henhold til en bestemt policy den følger, derav fairness, priority, responsiveness osv.

Scheduleren bruker timere, context switches og kjørekøer for å stanse og gjenoppta threads. Dette skaper en illusjon av at flere oppgaver kjører samtidig. 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> Threads brukes for å muliggjøre parallelisme på multi-core systemer, men også for å strukturere programmer og for å gjøre brukergrensesnitt mer responsive. For eksempel kan programmer struktureres slik at uavhengige oppgaver kjøres i separate threads, noe som kan gjøre koden mer oversiktlig. I brukergrensesnitt kan man bruke en egen UI-thread for å håndtere input, mens tidkrevende arbeid utføres i bakgrunnsthreads, slik at applikasjonen forblir responsiv. Scheduleren prioriterer UI-thread, slik at brukerinput behandles raskt. 

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Fibers og coroutines er lette alternativer til threads som styres av programmet selv i stedet for av operativsystemet. De krever mindre ressurser, er raskere å bytte mellom og gjør det enklere å unngå kompliserte synkroniseringsfeil. 

Vi velger ofte disse når vi trenger enorm samtidighet – som å håndtere tusenvis eller millioner av oppgaver samtidig – spesielt i programmer som venter mye på inn- og utdata slik som webservere. 

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Begge deler. Det kan forenkle strukturen til et program ved å naturlige dele inn uavhengige oppgaver inn i threads, men det introduserer kompleksitet i form av synkronisering, race conditions og debugging-vansker. Så concurrency kan forbedre ytelse og responsivhet, men det er vanskeligere å jobbe med og man risikerer bugs som er vanskelige å oppdage.

What do you think is best - *shared variables* or *message passing*?
> Message passing er det ryddigste, ettersom det eliminerer data races. Shared variables er (i øvingen ihvertfall) enkelt å forstå og bruke på enkle oppgaver, men jeg vil tro at det blir mye å holde styr på når oppgavestørrelsen vokser.


