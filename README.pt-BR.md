# Equilotl Portable — Discord Lite Support

[English](README.md) · [Instalação detalhada](docs/INSTALLATION.md) · [Solução de problemas](docs/TROUBLESHOOTING.md)

Fork comunitário independente do [Equilotl](https://github.com/Equicord/Equilotl), baseado na versão **2.3.0**, com suporte ao Discord Lite portátil da Team AntiLag.

**Não é uma versão oficial.** Não possui afiliação com Discord, afiliação oficial com Equicord ou afiliação com Team AntiLag. Modificações do cliente podem estar sujeitas às [regras do Discord](https://discord.com/terms).

## Recursos

- Descoberta automática e seleção manual de instalações válidas.
- Instalação, reparação/atualização do Equicord e desinstalação pela GUI/CLI.
- Suporte a `resources/app` preservando o bootstrap AntiLag, executáveis e DLLs.
- Backups verificados, rollback, recuperação e diagnóstico de processos por caminho/PID.
- Fluxo de ASAR do Discord oficial mantido. OpenAsar não é compatível com o formato portátil descompactado.

## Baixar e usar

Baixe da [página de Releases deste fork](https://github.com/GordinTv/equilotl-portable/releases). Enquanto não houver arquivos publicados nessa página, não há download público anunciado. Os downloads do instalador original não contêm estas alterações.

1. Extraia o ZIP Windows x64 e confira `SHA256SUMS.txt`.
2. Feche o Lite completamente, inclusive pela bandeja.
3. Execute `Equilotl-Portable-windows-amd64.exe`.
4. Selecione a versão/caminho corretos e clique em **Install**.
5. Abra o Lite normalmente e procure as configurações do Equicord.
6. Use **Reinstall / Repair** para reparar/atualizar; **Uninstall** restaura o manifesto original.

```powershell
.\Equilotl-Portable-CLI-windows-amd64.exe --diagnose-processes --location "$env:APPDATA\DC-Antilag\app\app-1.0.1160"
.\Equilotl-Portable-CLI-windows-amd64.exe --repair --location "$env:APPDATA\DC-Antilag\app\app-1.0.1160"
.\Equilotl-Portable-CLI-windows-amd64.exe --uninstall --location "$env:APPDATA\DC-Antilag\app\app-1.0.1160"
```

O pacote não inclui Discord Lite, Equicord ou dados da conta. A desinstalação não apaga configurações do Equicord. Preserve os backups em caso de falha. Binários não são assinados, salvo indicação explícita na release. Hashes não comprovam identidade do publicador.

## Compatibilidade e limites

Distribuição para **Windows x64**. Há ensaios registrados com Lite PTB **1.0.1160** e PTB oficial **1.0.1230**, além de fixtures para outros canais. Não há garantia universal.

Os registros confirmam carregamento antes do login, dois plugins, gravação de configuração e restauração de 1.292 arquivos. O funcionamento após login foi posteriormente **relatado pelo mantenedor**, sem publicação de dados privados. Temas, voz, compartilhamento de tela e a checklist completa permanecem pendentes de registro. Veja [testes](docs/TESTING.md).

Um ensaio posterior antes do login carregou Equicord/NoTrack, mas falhou ao confirmar o início do FixDiscordCss; a validação atual dos plugins continua incompleta.

Permissões ou arquivos somente leitura podem impedir o probe de segurança sem significar que o Lite está aberto. Não desative proteções para contornar erros. O auto-update do instalador para o upstream foi desativado; a atualização do Equicord por Repair continua disponível.

## Compilar e contribuir

Requer Go **1.26.10**, Windows x64, w64devkit fixado por checksum, Python 3.11+ para empacotar e Node 24 para o teste do loader.

```powershell
.\scripts\Setup-Toolchain.ps1
.\scripts\Test-Windows.ps1
.\scripts\Build-Windows.ps1
python scripts/package_release.py
```

[Compilação](docs/BUILDING.md) · [Contribuição](CONTRIBUTING.md) · [Histórico](CHANGELOG.md) · [Segurança](SECURITY.md). Os guias técnicos vinculados estão em inglês.

Use [Issues](https://github.com/GordinTv/equilotl-portable/issues) com erros sanitizados. Não envie tokens, cookies, pastas `data` ou logs sem revisão. Vulnerabilidades devem ser relatadas de forma privada.

## Licença

GNU GPLv3, preservando os créditos e termos originais. Veja [LICENSE](LICENSE), [NOTICE](NOTICE.md) e [dependências](THIRD_PARTY_NOTICES.md). Créditos ao Equilotl/Equicord, Vencord Installer, Vendicated, thororen1234 e demais contribuidores. Não há concessão de direitos de marca nem alegação de endosso.
