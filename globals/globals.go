package globals

import tea "github.com/charmbracelet/bubbletea"

var program *tea.Program

func InitProgram(p *tea.Program) {
    program = p
}

func GetProgram() *tea.Program {
    return program
}
