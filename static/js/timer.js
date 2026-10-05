// Shared lobby turn/round countdown for gameshell games.
//
// The countdown is entirely client-side: the server stores a seconds value
// (LOBBY_SETTINGS.TURN_TIMER_SECONDS) and broadcasts hints; each browser runs
// its own interval. What happens at zero is game-specific, so it is supplied
// as a callback rather than baked in here. Games mount the framework static
// assets (e.g. under /gs/) and drive this from their websocket handler.
window.gsTimer = (function () {
    var interval = null;
    var element = null;

    // Class thresholds, in seconds remaining.
    var WARN_AT = 10;
    var CRITICAL_AT = 5;

    function render(secondsRemaining) {
        if (!element) return;
        element.textContent = secondsRemaining + "s";
        element.classList.remove("warn", "critical");
        if (secondsRemaining <= CRITICAL_AT) {
            element.classList.add("critical");
        } else if (secondsRemaining <= WARN_AT) {
            element.classList.add("warn");
        }
    }

    // stop halts any running countdown and leaves the display as-is.
    function stop() {
        if (interval) {
            clearInterval(interval);
            interval = null;
        }
    }

    // reset stops the countdown and shows a static value (or blanks the
    // display when seconds is 0), without starting anything.
    function reset(el, seconds) {
        stop();
        element = el || element;
        if (!element) return;
        element.classList.remove("warn", "critical");
        element.textContent = seconds ? seconds + "s" : "";
    }

    // start begins a fresh countdown on el from seconds, invoking onExpire
    // once when it reaches zero. Passing 0 (or no) seconds just clears the
    // display, so callers can treat "timer off" as an ordinary start call.
    function start(el, seconds, onExpire) {
        reset(el, seconds);
        if (!element || !seconds || seconds <= 0) return;

        var secondsRemaining = seconds;
        render(secondsRemaining);

        interval = setInterval(function () {
            secondsRemaining -= 1;
            if (secondsRemaining < 0) secondsRemaining = 0;
            render(secondsRemaining);

            if (secondsRemaining === 0) {
                stop();
                if (typeof onExpire === "function") onExpire();
            }
        }, 1000);
    }

    // isRunning reports whether a countdown is currently ticking, so callers
    // can avoid restarting one on every incidental refresh.
    function isRunning() {
        return interval !== null;
    }

    return { start: start, stop: stop, reset: reset, isRunning: isRunning };
})();
