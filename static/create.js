(function () {
  const form = document.getElementById("create-form");
  const textEl = document.getElementById("text");
  const ttlEl = document.getElementById("ttl");
  const submitBtn = document.getElementById("submit-btn");
  const result = document.getElementById("result");
  const shareLink = document.getElementById("share-link");
  const copyBtn = document.getElementById("copy-btn");
  const newNoteBtn = document.getElementById("new-note-btn");
  const errorEl = document.getElementById("error");
  const pill = document.getElementById("main-pill");
  const chips = document.getElementById("ttl-chips");

  function showError(msg) {
    errorEl.textContent = msg;
    errorEl.classList.remove("hidden");
  }

  function hideError() {
    errorEl.classList.add("hidden");
    errorEl.textContent = "";
  }

  function setLoading(btn, on) {
    btn.disabled = on;
    btn.classList.toggle("loading", on);
  }

  // Tabs
  document.querySelectorAll(".seg-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      const tab = btn.dataset.tab;
      document.querySelectorAll(".seg-btn").forEach((b) => {
        const on = b === btn;
        b.classList.toggle("active", on);
        b.setAttribute("aria-selected", on ? "true" : "false");
      });
      pill.classList.toggle("to-about", tab === "about");

      document.querySelectorAll(".tab-panel").forEach((panel) => {
        const on = panel.id === `tab-${tab}`;
        panel.classList.toggle("active", on);
        panel.hidden = !on;
      });
    });
  });

  // TTL chips
  chips.addEventListener("click", (e) => {
    const chip = e.target.closest(".chip");
    if (!chip) return;
    chips.querySelectorAll(".chip").forEach((c) => {
      const on = c === chip;
      c.classList.toggle("active", on);
      c.setAttribute("aria-checked", on ? "true" : "false");
    });
    ttlEl.value = chip.dataset.ttl;
  });

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    hideError();
    result.classList.add("hidden");

    const plaintext = textEl.value;
    if (!plaintext.trim()) {
      showError("Введите текст записки.");
      return;
    }
    if (new TextEncoder().encode(plaintext).length > 8192) {
      showError("Записка слишком большая (макс. ~8 KB текста).");
      return;
    }

    setLoading(submitBtn, true);
    try {
      const key = await QNotesCrypto.generateKey();
      const ciphertext = await QNotesCrypto.encryptText(plaintext, key.cryptoKey);

      const res = await fetch("/api/notes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ciphertext, ttl: ttlEl.value }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || "не удалось создать записку");
      }

      const data = await res.json();
      const url = `${location.origin}/note/${data.id}#${key.hex}`;
      shareLink.value = url;
      result.classList.remove("hidden");
      textEl.value = "";
      result.scrollIntoView({ behavior: "smooth", block: "nearest" });
    } catch (err) {
      showError(err.message || "Ошибка создания записки");
    } finally {
      setLoading(submitBtn, false);
    }
  });

  copyBtn.addEventListener("click", async () => {
    try {
      await navigator.clipboard.writeText(shareLink.value);
      const prev = copyBtn.textContent;
      copyBtn.textContent = "Готово";
      setTimeout(() => {
        copyBtn.textContent = prev;
      }, 1400);
    } catch {
      shareLink.select();
      document.execCommand("copy");
    }
  });

  newNoteBtn.addEventListener("click", () => {
    result.classList.add("hidden");
    shareLink.value = "";
    textEl.focus();
  });
})();
