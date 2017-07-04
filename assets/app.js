var mood = {}

$(function(){

	if (!window["WebSocket"]) {
		alert('Mood by Machine Box requires web sockets, open this in a modern web browser.')
		return
	}

	var socket = new WebSocket('ws://'+location.host+'/analysis');
	socket.onclose = function() {
		mood.socketOpen = false
    	console.info('connection closed')
    }
	socket.onmessage = function(e) {
		var obj = JSON.parse(e.data)
		mood.updateView(obj)
	}
	socket.onopen = function(){
		mood.socketOpen = true
		mood.updateSettings()
	}

	$('#add-button').click(function(){
		var grid = $('.ui.term.grid')
		$('#term-template').clone().attr('id','').show().appendTo(grid).find('input').focus()
		grid.find('.empty.column').appendTo(grid)
	})
	$(document).on('submit', 'form.term', function(e){
		e.preventDefault()
		var $this = $(this)
		$this.attr('data-term', $this.find('input[name=term]').val())
		mood.updateSettings()
		$this.addClass('loading')
	})
	$(document).on('click', '.remove.button', function(e){
		e.preventDefault()
		console.info($(e.target).closest('.column'))
		$(e.target).closest('.term.form').css({overflow:'hidden'}).animate({ width: 0 }, {
			complete: function(){
				$(this).closest('.column').remove()
				mood.updateSettings()
			}
		})
	})

	// mood.updateSettings updates the analysis settings by sending
	// the object through the web socket to the server.
	mood.updateSettings = function() {
		var terms = []
		$('.ui.term.grid').find('input[name="term"]').each(function(){
			terms.push($(this).val())
		})
		console.info('Updating terms:', terms)
		socket.send(JSON.stringify({
			terms: terms
		}))
	}

	mood.updateView = function(data) {
		if (data.error) {
			$('.ui.error.message').text(data.error).show()
			return
		}
		$('.ui.error.message').hide()

		for (var term in data.tally) {
			if (!data.tally.hasOwnProperty(term)) { continue }
			var thisTally = data.tally[term]
			console.info(term, thisTally)
			var form = $('form[data-term="'+term+'"]')
			form.removeClass('loading')
			form.find('.details').show()

			// sentiment and count
			var sentimentIcon = 'meh'
			if (thisTally.sentiment_average >= 0.6) {
				sentimentIcon = 'smile'
			} else if (thisTally.sentiment_average <= 0.4) {
				sentimentIcon = 'frown'
			}

			form.find('[data-content="count"]').text(thisTally.count)
			form.find('[data-content="sentiment"]').text(thisTally.sentiment_average)
			form.find('[data-content="sentiment-overview"]').empty().append(
				$("<i>", {'class': sentimentIcon+' icon'})
			)

			// keywords
			var keywordsEl = form.find('.keywords').empty()
			for (var k in thisTally.top_keywords) {
				if (!thisTally.top_keywords.hasOwnProperty(k)) { continue }
				var keyword = thisTally.top_keywords[k]
				keywordsEl.append($("<li>").text(keyword.keyword))
			}

			// entities
			for (var entityType in thisTally.top_entities) {
				if (!thisTally.top_entities.hasOwnProperty(entityType)) { continue }
				var entities = thisTally.top_entities[entityType]

				var entitiesEl = form.find('[data-content="'+entityType+'-entities"]')
				if (entitiesEl.length === 0) {
					entitiesEl = $('<div>').appendTo(form.find('[data-content="entities"]')).attr('data-content', entityType+'-entities')
					entitiesEl.append(
						$("<h2>").text(entityType),
						$("<ul>")
					)
				}
				var entitiesListEl = entitiesEl.find('ul').empty()

				for (var k in entities) {
					if (!entities.hasOwnProperty(k)) { continue }
					var entity = entities[k]
					entitiesListEl.append(
						$('<li>').append(
							$('<div>', {class:'ui label'}).append(
								entity.text,
								$('<span>', {class:'detail'}).text('x'+entity.count)
							)
						)
					)
				}

			}
		}
	}

})
