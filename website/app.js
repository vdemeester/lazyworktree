// Theme
(function () {
  const root = document.documentElement;
  const toggle = document.querySelector(".theme-toggle");
  const prefersDark = window.matchMedia("(prefers-color-scheme: dark)");

  function applyTheme(theme) {
    root.dataset.theme = theme;
  }

  function getEffectiveTheme() {
    const stored = localStorage.getItem("theme");
    if (stored) return stored;
    return prefersDark.matches ? "dark" : "light";
  }

  applyTheme(getEffectiveTheme());

  if (toggle) {
    toggle.addEventListener("click", function () {
      var next = root.dataset.theme === "dark" ? "light" : "dark";
      localStorage.setItem("theme", next);
      applyTheme(next);
    });
  }

  prefersDark.addEventListener("change", function () {
    if (!localStorage.getItem("theme")) {
      applyTheme(prefersDark.matches ? "dark" : "light");
    }
  });
})();

const revealNodes = document.querySelectorAll(".reveal");

if ("IntersectionObserver" in window) {
  const revealObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("visible");
          revealObserver.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.15 }
  );

  revealNodes.forEach((node, index) => {
    node.style.transitionDelay = `${Math.min(index * 80, 350)}ms`;
    revealObserver.observe(node);
  });
} else {
  revealNodes.forEach((node) => {
    node.classList.add("visible");
  });
}

const tabButtons = document.querySelectorAll(".tab-btn");
const tabPanels = document.querySelectorAll(".tab-panel");

tabButtons.forEach((button) => {
  button.addEventListener("click", () => {
    tabButtons.forEach((item) => {
      item.classList.remove("active");
      item.setAttribute("aria-selected", "false");
    });
    tabPanels.forEach((panel) => {
      panel.classList.remove("active");
      panel.hidden = true;
    });

    button.classList.add("active");
    button.setAttribute("aria-selected", "true");
    const panel = document.getElementById(button.dataset.target);
    if (!panel) {
      return;
    }
    panel.classList.add("active");
    panel.hidden = false;
  });
});

// Lightbox
const lightbox = document.getElementById("lightbox");
const lightboxImg = document.getElementById("lightbox-img");

function openLightbox(src, alt) {
  lightboxImg.src = src;
  lightboxImg.alt = alt;
  lightboxImg.classList.remove("zoomed");
  lightbox.hidden = false;
  document.body.style.overflow = "hidden";
}

function closeLightbox() {
  lightbox.hidden = true;
  lightboxImg.src = "";
  document.body.style.overflow = "";
}

document.querySelectorAll(".preview-card img").forEach((img) => {
  img.addEventListener("click", () => {
    openLightbox(img.src, img.alt);
  });
});

lightbox.addEventListener("click", (e) => {
  if (e.target === lightboxImg) {
    lightboxImg.classList.toggle("zoomed");
  } else {
    closeLightbox();
  }
});

document.addEventListener("keydown", (e) => {
  if (e.key === "Escape" && !lightbox.hidden) {
    closeLightbox();
  }
});
