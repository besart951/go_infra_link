<!--
  AnimatedAirplane.svelte
  Multiple small line-art airplanes flying across a sky background.
  Pure SVG + CSS — no JavaScript, no external dependencies.
  Supports light/dark mode via the existing .dark class on <html>.

  Colors are applied via CSS (not inline SVG attributes) so they can
  respond to .dark mode using the project's design-token pattern.
-->

<svg
  class="airplane-scene"
  xmlns="http://www.w3.org/2000/svg"
  viewBox="0 0 800 500"
  preserveAspectRatio="xMidYMid slice"
  aria-hidden="true"
>
  <!-- ===== Sky background ===== -->
  <rect class="sky" width="800" height="500" />

  <!-- ===== Decorative clouds ===== -->
  <g class="cloud cloud-1">
    <ellipse cx="120" cy="100" rx="55" ry="18" />
    <ellipse cx="85" cy="95" rx="35" ry="14" />
    <ellipse cx="160" cy="96" rx="40" ry="13" />
  </g>
  <g class="cloud cloud-2">
    <ellipse cx="550" cy="80" rx="65" ry="20" />
    <ellipse cx="510" cy="74" rx="38" ry="15" />
    <ellipse cx="600" cy="76" rx="42" ry="14" />
  </g>
  <g class="cloud cloud-3">
    <ellipse cx="350" cy="380" rx="70" ry="16" />
    <ellipse cx="300" cy="375" rx="42" ry="13" />
    <ellipse cx="410" cy="377" rx="45" ry="12" />
  </g>
  <g class="cloud cloud-4">
    <ellipse cx="700" cy="300" rx="50" ry="14" />
    <ellipse cx="670" cy="296" rx="30" ry="11" />
    <ellipse cx="740" cy="298" rx="35" ry="10" />
  </g>
  <g class="cloud cloud-5">
    <ellipse cx="200" cy="440" rx="60" ry="14" />
    <ellipse cx="165" cy="436" rx="36" ry="11" />
    <ellipse cx="245" cy="438" rx="38" ry="10" />
  </g>

  <!-- ===== AIRPLANES ===== -->

  <!-- ── Plane 1: left→right, top area, medium ── -->
  <g class="fly-ltr fly-ltr-1">
    <g class="bob bob-1">
      <g transform="translate(0, 90) scale(1.8)" class="plane plane-a">
        <path d="M10,20h2.11a1,1,0,0,0,.81-.42L16.84,14H19a2,2,0,0,0,2-2h0a2,2,0,0,0-2-2H16.84L12.92,4.42A1,1,0,0,0,12.11,4H10l1,6H8L6.29,8.29A1,1,0,0,0,5.59,8H3l2,4L3,16H5.59a1,1,0,0,0,.7-.29L8,14h3Z" />
      </g>
    </g>
  </g>

  <!-- ── Plane 2: right→left, middle area, larger ── -->
  <g class="fly-rtl fly-rtl-1">
    <g class="bob bob-2">
      <g transform="translate(0, 200) scale(-2.4, 2.4)" class="plane plane-b">
        <path d="M10,20h2.11a1,1,0,0,0,.81-.42L16.84,14H19a2,2,0,0,0,2-2h0a2,2,0,0,0-2-2H16.84L12.92,4.42A1,1,0,0,0,12.11,4H10l1,6H8L6.29,8.29A1,1,0,0,0,5.59,8H3l2,4L3,16H5.59a1,1,0,0,0,.7-.29L8,14h3Z" />
      </g>
    </g>
  </g>

  <!-- ── Plane 3: left→right, lower area, small ── -->
  <g class="fly-ltr fly-ltr-2">
    <g class="bob bob-3">
      <g transform="translate(0, 330) scale(1.3)" class="plane plane-c">
        <path d="M10,20h2.11a1,1,0,0,0,.81-.42L16.84,14H19a2,2,0,0,0,2-2h0a2,2,0,0,0-2-2H16.84L12.92,4.42A1,1,0,0,0,12.11,4H10l1,6H8L6.29,8.29A1,1,0,0,0,5.59,8H3l2,4L3,16H5.59a1,1,0,0,0,.7-.29L8,14h3Z" />
      </g>
    </g>
  </g>

  <!-- ── Plane 4: right→left, top-right, tiny ── -->
  <g class="fly-rtl fly-rtl-2">
    <g class="bob bob-4">
      <g transform="translate(0, 50) scale(-1, 1)" class="plane plane-d">
        <path d="M10,20h2.11a1,1,0,0,0,.81-.42L16.84,14H19a2,2,0,0,0,2-2h0a2,2,0,0,0-2-2H16.84L12.92,4.42A1,1,0,0,0,12.11,4H10l1,6H8L6.29,8.29A1,1,0,0,0,5.59,8H3l2,4L3,16H5.59a1,1,0,0,0,.7-.29L8,14h3Z" />
      </g>
    </g>
  </g>

  <!-- ── Plane 5: left→right, mid-low, medium ── -->
  <g class="fly-ltr fly-ltr-3">
    <g class="bob bob-5">
      <g transform="translate(0, 260) scale(1.6)" class="plane plane-e">
        <path d="M10,20h2.11a1,1,0,0,0,.81-.42L16.84,14H19a2,2,0,0,0,2-2h0a2,2,0,0,0-2-2H16.84L12.92,4.42A1,1,0,0,0,12.11,4H10l1,6H8L6.29,8.29A1,1,0,0,0,5.59,8H3l2,4L3,16H5.59a1,1,0,0,0,.7-.29L8,14h3Z" />
      </g>
    </g>
  </g>

  <!-- ── Plane 6: right→left, bottom, small ── -->
  <g class="fly-rtl fly-rtl-3">
    <g class="bob bob-6">
      <g transform="translate(0, 410) scale(-1.2, 1.2)" class="plane plane-f">
        <path d="M10,20h2.11a1,1,0,0,0,.81-.42L16.84,14H19a2,2,0,0,0,2-2h0a2,2,0,0,0-2-2H16.84L12.92,4.42A1,1,0,0,0,12.11,4H10l1,6H8L6.29,8.29A1,1,0,0,0,5.59,8H3l2,4L3,16H5.59a1,1,0,0,0,.7-.29L8,14h3Z" />
      </g>
    </g>
  </g>

</svg>

<style>
  /* ===========================
     Container
     =========================== */
  .airplane-scene {
    width: 100%;
    height: 100%;
    display: block;
  }

  /* ===========================
     LIGHT MODE colors (default)
     =========================== */

  /* Sky background */
  .sky {
    fill: var(--color-muted);
  }

  /* Clouds */
  .cloud ellipse {
    fill: white;
  }
  .cloud-1 { opacity: 0.55; }
  .cloud-2 { opacity: 0.45; }
  .cloud-3 { opacity: 0.35; }
  .cloud-4 { opacity: 0.30; }
  .cloud-5 { opacity: 0.25; }

  /* Airplane paths — line art */
  .plane path {
    fill: none;
    stroke: var(--color-foreground);
    stroke-width: 1.4;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  /* Varying opacities per plane for depth */
  .plane-a { opacity: 0.35; }
  .plane-b { opacity: 0.20; }
  .plane-c { opacity: 0.18; }
  .plane-d { opacity: 0.12; }
  .plane-e { opacity: 0.28; }
  .plane-f { opacity: 0.15; }

  /* ===========================
     DARK MODE overrides
     Uses the .dark class from the project's theme system.
     The @custom-variant dark(&:is(.dark *)) means
     :global(.dark) selectors will work.
     =========================== */

  :global(.dark) .sky {
    fill: var(--sidebar);
  }

  :global(.dark) .cloud ellipse {
    fill: var(--muted);
  }

  :global(.dark) .cloud-1 { opacity: 0.3; }
  :global(.dark) .cloud-2 { opacity: 0.22; }
  :global(.dark) .cloud-3 { opacity: 0.18; }
  :global(.dark) .cloud-4 { opacity: 0.15; }
  :global(.dark) .cloud-5 { opacity: 0.12; }

  :global(.dark) .plane path {
    stroke: var(--foreground);
  }

  :global(.dark) .plane-a { opacity: 0.22; }
  :global(.dark) .plane-b { opacity: 0.14; }
  :global(.dark) .plane-c { opacity: 0.12; }
  :global(.dark) .plane-d { opacity: 0.08; }
  :global(.dark) .plane-e { opacity: 0.18; }
  :global(.dark) .plane-f { opacity: 0.10; }

  /* ===========================
     Left-to-right flight paths
     =========================== */
  .fly-ltr-1 { animation: ltr 28s linear infinite; }
  .fly-ltr-2 { animation: ltr 35s linear infinite; animation-delay: -12s; }
  .fly-ltr-3 { animation: ltr 22s linear infinite; animation-delay: -6s; }

  @keyframes ltr {
    0%   { transform: translateX(-80px); }
    100% { transform: translateX(850px); }
  }

  /* ===========================
     Right-to-left flight paths
     =========================== */
  .fly-rtl-1 { animation: rtl 32s linear infinite; animation-delay: -4s; }
  .fly-rtl-2 { animation: rtl 40s linear infinite; animation-delay: -18s; }
  .fly-rtl-3 { animation: rtl 26s linear infinite; animation-delay: -10s; }

  @keyframes rtl {
    0%   { transform: translateX(850px); }
    100% { transform: translateX(-80px); }
  }

  /* ===========================
     Gentle vertical bobbing
     =========================== */
  .bob-1 { animation: bob-a 5s ease-in-out infinite; }
  .bob-2 { animation: bob-b 6s ease-in-out infinite; animation-delay: -2s; }
  .bob-3 { animation: bob-a 4.5s ease-in-out infinite; animation-delay: -1s; }
  .bob-4 { animation: bob-b 7s ease-in-out infinite; animation-delay: -3s; }
  .bob-5 { animation: bob-a 5.5s ease-in-out infinite; animation-delay: -4s; }
  .bob-6 { animation: bob-b 4s ease-in-out infinite; animation-delay: -1.5s; }

  @keyframes bob-a {
    0%, 100% { transform: translateY(0); }
    50%      { transform: translateY(-10px); }
  }
  @keyframes bob-b {
    0%, 100% { transform: translateY(0); }
    50%      { transform: translateY(8px); }
  }

  /* ===========================
     Cloud drift
     =========================== */
  .cloud-1 { animation: drift1 30s ease-in-out infinite; }
  .cloud-2 { animation: drift2 38s ease-in-out infinite; }
  .cloud-3 { animation: drift3 42s ease-in-out infinite; }
  .cloud-4 { animation: drift2 34s ease-in-out infinite reverse; }
  .cloud-5 { animation: drift1 36s ease-in-out infinite reverse; }

  @keyframes drift1 {
    0%, 100% { transform: translateX(0); }
    50%      { transform: translateX(60px); }
  }
  @keyframes drift2 {
    0%, 100% { transform: translateX(0); }
    50%      { transform: translateX(-50px); }
  }
  @keyframes drift3 {
    0%, 100% { transform: translateX(0); }
    50%      { transform: translateX(40px); }
  }
</style>
