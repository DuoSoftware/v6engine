/*jsPlumb.ready(function() {

    jsPlumb.setContainer($('#container'));
  
    var i = 0;

    $('#container').dblclick(function(e) {
    var newState = $('<div>').attr('id', 'state' + i).addClass('item');
    
    var title = $('<div>').addClass('title').text('State ' + i);
    var connect = $('<div>').addClass('connect');
    //var hoverMessage = $('<p>').addClass('hoverMessage').text('Connection drop location...');
    
    newState.css({
      'top': e.pageY,
      'left': e.pageX
    });
    
    newState.append(title);
    //connect.append(hoverMessage);
    newState.append(connect);
    
    
    $('#container').append(newState);
    
    jsPlumb.makeTarget(newState, {
      endpoint:"Blank",
      anchor:[ "Perimeter", { shape:"Square", anchorCount:8 }],
      connectorOverlays:[ 
        [ "Arrow", { width:30, length:30, location:1, id:"arrow" } ]
      ]
    });
    
    jsPlumb.makeSource(connect, {
      parent: newState,
      endpoint:"Blank",
      anchor:[ "Perimeter", { shape:"Square", anchorCount:8 }],
      connectorOverlays:[ 
        [ "Arrow", { width:30, length:30, location:1, id:"arrow" } ]
      ]
    });
    
    jsPlumb.draggable(newState, {
      containment: 'parent'
    });

    newState.dblclick(function(e) {
      jsPlumb.detachAllConnections($(this));
      $(this).remove();
      e.stopPropagation();
    });

    newState.hover(function(e){

    });
    
    i++;    
    });  
  });*/

/*jsPlumb.importDefaults({
  PaintStyle : {
    lineWidth:13,
    strokeStyle: 'rgba(200,0,0,0.5)'
  },
  DragOptions : { cursor: "crosshair" },
  Endpoints : [ [ "Dot", { radius:7 } ], [ "Dot", { radius:25 } ] ],
  EndpointStyles : [{ fillStyle:"#225588" }, { fillStyle:"#558822" }],
  
});*/