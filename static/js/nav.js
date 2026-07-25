// Shared click-to-navigate behavior for gameshell games.
//
// Elements that navigate on click (buttons, top-bar menu rows) are not links,
// so the browser gives them none of a link's affordances — ctrl/cmd-click,
// middle-click, and shift-click all just navigate in place. Marking them with
// data-href instead of an onclick handler routes them through here, which
// honors those modifiers the way a real anchor would:
//
//   <button data-href="/decks">Card Decks</button>
//
// Genuine <a href> elements need nothing from this file; the browser already
// handles them. Use data-href only where an anchor is awkward.
window.gsNav = (function () {
    // openInNewTab reports whether this click asked for a new tab/window,
    // matching what a browser does with a real link.
    function wantsNewTab(event) {
        return event.ctrlKey || event.metaKey || event.shiftKey;
    }

    function target(event) {
        var el = event.target.closest("[data-href]");
        if (!el) return null;
        // A real link inside the element keeps its own behavior.
        if (event.target.closest("a[href]")) return null;
        return el;
    }

    function navigate(el, newTab) {
        var href = el.getAttribute("data-href");
        if (!href) return;
        if (newTab) {
            // No features string: passing one (even "noopener") makes some
            // browsers open a popup window instead of a tab, and popup
            // blockers treat it more harshly. Sever the opener afterwards
            // instead, which gets the same protection with tab behavior.
            var opened = window.open(href, "_blank");
            if (opened) opened.opener = null;
        } else {
            window.location.href = href;
        }
    }

    // Left click, with or without a modifier.
    document.addEventListener("click", function (event) {
        var el = target(event);
        if (!el) return;
        event.preventDefault();
        navigate(el, wantsNewTab(event));
    });

    // Middle click. Fires as auxclick; button 1 is the middle button.
    document.addEventListener("auxclick", function (event) {
        if (event.button !== 1) return;
        var el = target(event);
        if (!el) return;
        event.preventDefault();
        navigate(el, true);
    });

    // Keyboard activation, so a data-href element that is not a <button>
    // (which already fires click on Enter/Space) is still reachable.
    document.addEventListener("keydown", function (event) {
        if (event.key !== "Enter" && event.key !== " ") return;
        var el = event.target.closest && event.target.closest("[data-href]");
        if (!el || el.tagName === "BUTTON" || el.tagName === "A") return;
        event.preventDefault();
        navigate(el, wantsNewTab(event));
    });

    return { navigate: navigate, wantsNewTab: wantsNewTab };
})();
