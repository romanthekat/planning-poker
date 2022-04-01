const svgFart = document.querySelector('.fart-frames-svg')
const fartFrames = svgFart.querySelectorAll('[class*="fart-frame-"]')
const copyButton = document.querySelector('.copy-link-button')
const logo = document.querySelector('#logo')
const sessionId = document.querySelector('#session-id-copy')

logo.addEventListener('click', _ => {
	gsap.to(fartFrames, {keyframes: [{opacity:1}, {opacity:0}], stagger:0.45, duration: 1.2, ease: "power1.out"})
})

copyButton.addEventListener('click', _ => {
	gsap.timeline()
		.to(sessionId, {opacity:1, duration:0.01})
		.to(sessionId, {opacity:0, scale: 2, duration:1})
		.to(sessionId, {scale: 1})
})