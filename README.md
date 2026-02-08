# orbiqd-projectkit

[![codecov](https://codecov.io/gh/orbiqd/orbiqd-projectkit/branch/main/graph/badge.svg)](https://app.codecov.io/gh/orbiqd/orbiqd-projectkit)

- projectkit ma na celu ulatwienie zarzadzania instrukcjami ai, skillami czy worklowami.
- projectkit rowniez zarzadza dokumentacja projektu w formie standardow oraz design docow
- projectkit umozliwia reuzywanie tych specyfikacji w innych projektach
- projectkit renderuje standardy projektu do plikow w formie markdownu jako dokumentacji
- projectkit renderuje pliki z instrukcjami dla agentow w formacie specyficznym dla kazdego agenta
- projectkit udostepnia serwer mcp ktory pozwala na podazanie za konkretnym workflow
- workflow to zestaw regul pracy ktore agent musi stosowac, aby wykoanc jakies okreslone zadanie
- paczka zestawu ai workflow, skill oraz instrukcji a takze standardow czy design docow nazywana jest rulebook
- rulebook moze byc pobrany z zewnatrz, moze tez istniec w katalogu projektu
- konfguracja projectkit - czuyli skad wziac rulebooki lub pojedyncze konfiguracje instrukcji/standardow znajduje sie w roocie projektu w pliku .projectkit.yaml
- plik z konfiguracja ma byc wyszukiwany w aktualnie uruchamianym katalogu ale takze w home usera. jesli sa znalezione dwa, to ten z projektu ma priorytet, ale sa one mergowane
- rulebooki maja swoja konfiguracje w pliku rulebook.yaml w katalogu z rulebookiem
- rulebooki oraz pojedyncze konfiguracje mozna pobierac z lokalnego storage lub z git
