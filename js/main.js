
/* 1. Mostrar elementos al hacer scroll */
const revealObserver = new IntersectionObserver(
    entries => entries.forEach(entry => {
        if (entry.isIntersecting) {
            entry.target.classList.add('visible');
            revealObserver.unobserve(entry.target);
        }
    }),
    { threshold: 0.08, rootMargin: '0px 0px -30px 0px' }
);
document.querySelectorAll('.reveal').forEach(el => revealObserver.observe(el));

/* 2. Sombra en la barra de navegación al hacer scroll */
const header = document.getElementById('navbar');
window.addEventListener('scroll', () => {
    header.classList.toggle('scrolled', window.scrollY > 30);
}, { passive: true });

/* 3. Enlace activo en la navegación */
const sections = document.querySelectorAll('section[id]');
const navLinks = document.querySelectorAll('.nav-links a');

const sectionObserver = new IntersectionObserver(
    entries => entries.forEach(entry => {
        if (entry.isIntersecting) {
            navLinks.forEach(l => l.classList.remove('active'));
            const link = document.querySelector(`.nav-links a[href="#${entry.target.id}"]`);
            if (link) link.classList.add('active');
        }
    }),
    { threshold: 0.35 }
);
sections.forEach(sec => sectionObserver.observe(sec));
