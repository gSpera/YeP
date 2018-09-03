function insertTab(event) {
    //If not Tab or Alt Key is pressed
    if (event.key != "Tab" || event.altKey) {
        return true
    }
    const insert = "\t"
    const value = code.value
    const before = value.substring(0, code.selectionStart)
    const after = value.substring(code.selectionEnd, code.value.length)
    code.value = before + insert + after
    code.selectionStart = code.selectionEnd = before.length + insert.length
    return false
}

window.onload = () => {
    let code = document.getElementById("code")
    code.onkeydown = insertTab
    code.focus()
}