# Third-party notices

Retain LICENSE, NOTICE.md and the entire `third_party` directory with binary redistribution. This directory contains verbatim notices collected from the locked Go dependency graph and the Go runtime. `third_party/modules.json` records module versions, notice paths and content hashes. Regenerate using `python scripts/collect_licenses.py` after `go mod download` and review the diff.

Main libraries include AllenDang/giu and Dear ImGui (MIT), Go OpenGL/GLFW bindings and GLFW (their retained permissive notices), ProtonMail/go-appdir (MIT), fatih/color (MIT), promptui/readline and their authors, and Go's x/image/x/sys packages (BSD-style notices). The inventory, not this short list, is the complete notice set for the locked module graph, including test/transitive modules.

Native Windows builds use w64devkit/GCC and MinGW-w64. The supplied `COPYING.MinGW-w64-runtime.txt` must accompany binaries; GCC runtime components use the GCC Runtime Library Exception. ImGui's embedded ProggyClean font identifies Tristan Grimmer as copyright holder under MIT; its notice is included separately. No Windows system DLL is bundled.

Permissive component notices do not replace the main project's GPL obligations. Distribute the exact corresponding source/build scripts alongside the binaries rather than pointing only at unmodified upstream. Preserve each component's own terms. This inventory is a redistribution review aid, not a claim that automated vulnerability/license scanners prove legal or security completeness.

Reference: [w64devkit licensing](https://github.com/skeeto/w64devkit#licenses), [GCC Runtime Library Exception](https://www.gnu.org/licenses/gcc-exception-3.1.en.html), [GPLv3](https://www.gnu.org/licenses/gpl-3.0.html).
