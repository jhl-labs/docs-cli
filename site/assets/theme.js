// Theme switch (light/dark/system) + copy-to-clipboard buttons.
// Applied before paint to avoid a flash, then wired up on DOMContentLoaded.
(function () {
  var KEY = "docs-cli-theme";
  function apply(choice) {
    var root = document.documentElement;
    if (choice === "light" || choice === "dark") {
      root.setAttribute("data-theme", choice);
    } else {
      root.removeAttribute("data-theme");
    }
  }
  try {
    apply(localStorage.getItem(KEY) || "system");
  } catch (e) {
    /* localStorage unavailable */
  }

  document.addEventListener("DOMContentLoaded", function () {
    var current = "system";
    try {
      current = localStorage.getItem(KEY) || "system";
    } catch (e) {}

    var buttons = document.querySelectorAll("[data-theme-choice]");
    function sync() {
      buttons.forEach(function (b) {
        b.setAttribute("aria-pressed", String(b.dataset.themeChoice === current));
      });
    }
    buttons.forEach(function (b) {
      b.addEventListener("click", function () {
        current = b.dataset.themeChoice;
        apply(current);
        try {
          localStorage.setItem(KEY, current);
        } catch (e) {}
        sync();
      });
    });
    sync();

    document.querySelectorAll("[data-copy-target]").forEach(function (btn) {
      btn.addEventListener("click", function () {
        var el = document.getElementById(btn.dataset.copyTarget);
        if (!el || !navigator.clipboard) return;
        navigator.clipboard.writeText(el.textContent.trim()).then(function () {
          var label = btn.dataset.copiedLabel || "Copied";
          var original = btn.dataset.copyLabel || btn.textContent;
          btn.textContent = label;
          setTimeout(function () {
            btn.textContent = original;
          }, 1500);
        });
      });
    });
  });
})();
