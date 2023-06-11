function mostrarMensaje(texto, duracion, tipo) {
	mensaje = $(`<div class='mensaje-emergente tipo-${tipo} '>${texto}</div>`)
	$(document.body).append(mensaje)
	
	setTimeout(() => {
		mensaje.addClass("mostrar");
	  }, 50);
	
	  setTimeout(() => {
		mensaje.addClass("esconder");
		setTimeout(() => {
			mensaje.remove();
		}, 500);
	  }, duracion);

	
}