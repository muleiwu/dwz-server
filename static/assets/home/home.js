(() => {
    const nav = document.querySelector('.home-nav');
    const toggle = nav?.querySelector('.home-nav-toggle');
    if (!toggle) return;

    const setOpen = (open) => {
        toggle.setAttribute('aria-expanded', String(open));
        toggle.setAttribute(
            'aria-label',
            open ? '关闭导航菜单' : '打开导航菜单',
        );
        nav.classList.toggle('is-open', open);
    };

    nav.classList.add('home-nav-enhanced');
    toggle.addEventListener('click', () =>
        setOpen(toggle.getAttribute('aria-expanded') !== 'true'),
    );
    nav.addEventListener('keydown', (event) => {
        if (
            event.key === 'Escape' &&
            toggle.getAttribute('aria-expanded') === 'true'
        ) {
            setOpen(false);
            toggle.focus();
        }
    });
    document.addEventListener('click', (event) => {
        if (!nav.contains(event.target)) setOpen(false);
    });
    nav.querySelectorAll('.home-nav-links a').forEach((link) => {
        link.addEventListener('click', () => setOpen(false));
    });
    window
        .matchMedia('(max-width: 640px)')
        .addEventListener('change', () => setOpen(false));
})();
