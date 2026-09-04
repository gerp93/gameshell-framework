// Shared lobby chat rendering for gameshell games.
//
// The framework's websocket broadcasts short control strings and human chat
// lines; system lines use inline color tokens (<blue>…</>, <green>…</>,
// <red>…</>). This renders one message into a chat-messages container, applies
// the color tokens, prepends a HH:MM timestamp, and trims history. Games mount
// the framework static assets (e.g. under /gs/) and call gsChat.append(...)
// from their own websocket onmessage handler.
window.gsChat = (function () {
    var MAX_MESSAGES = 100;

    function colorize(text) {
        return text
            .replaceAll("<red>", '<span class="gs-chat-red">')
            .replaceAll("<green>", '<span class="gs-chat-green">')
            .replaceAll("<blue>", '<span class="gs-chat-blue">')
            .replaceAll("</>", "</span>");
    }

    function timestamp() {
        var now = new Date();
        return (
            String(now.getHours()).padStart(2, "0") +
            ":" +
            String(now.getMinutes()).padStart(2, "0")
        );
    }

    // append renders one raw message line into the given container element.
    function append(messagesEl, rawText) {
        if (!messagesEl) return;
        var message = document.createElement("div");
        message.className = "gs-chat-message";
        message.innerHTML = timestamp() + " " + colorize(rawText);
        messagesEl.appendChild(message);
        while (messagesEl.childNodes.length > MAX_MESSAGES) {
            messagesEl.removeChild(messagesEl.childNodes[0]);
        }
        messagesEl.scrollTop = messagesEl.scrollHeight;
    }

    // wireForm wires a chat form so submitting sends the input over the socket
    // (short control strings never HTML) and clears it.
    function wireForm(formEl, inputEl, conn) {
        if (!formEl || !inputEl) return;
        formEl.onsubmit = function (event) {
            event.preventDefault();
            if (!inputEl.value) return;
            conn.send(inputEl.value);
            inputEl.value = "";
        };
    }

    return { append: append, wireForm: wireForm };
})();
