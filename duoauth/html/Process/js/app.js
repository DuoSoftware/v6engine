var app = angular.module('processdesigner',[]);

app.controller('mainController',['$scope','$http',function(scope,http){

    function module(library_id, schema_id, title, description, x, y, icon,variables) {
        this.library_id = library_id;
        this.schema_id = schema_id;
        this.title = title;
        this.description = description;
        this.x = x;
        this.y = y;
        this.icon = icon;
        this.variables = variables;
    }

    scope.library = [];
    scope.library_uuid = 0; 
    scope.schema = [];
    scope.schema_uuid = 0; 
    scope.library_topleft = {
            x: 15,
            y: 145,
            item_height: 50,
            margin: 5,
    };

    scope.module_css = {
            width: 150,
            height: 100, // actually variable
    };

    scope.variableslist = [];
    scope.variableslist =[
        {Key:'Attribute Details',Value:'INAttributeDetail'},
        {Key:'Attribute ID',Value:'Convert.ToString("INAttributeID")'},
        {Key:'Attribute Name',Value:'INName'},
        {Key:'Key',Value:'Value'},
        {Key:'Value',Value:'Convert.ToString("Key")'},
        {Key:'ID',Value:'Convert.ToInt64("4562187542124487")'},
        {Key:'Date',Value:'2014-12-10'}
    ];


    scope.openvariablebox = function(module){
        console.log("clicked element : "+ module.title);
        scope.vopenBox();
        scope.loadVariablesOfModule(module.variables);
    };

    scope.loadVariablesOfModule = function(variables){
        console.log("opened variables : " + variables);
        var obj = angular.fromJson(variables);

        scope.variableslist = [];

        angular.forEach(obj.Control.Variables, function(value, key) {
          scope.variableslist.push(key + ': ' + value);
        }, log);

        //scope.variableslist = obj.Control.Variables;

        alert(obj.Control.Name);
        console.log(scope.variableslist);
    };

    scope.redraw = function() {
        console.log("-- Redraw function executed");
        scope.schema_uuid = 0;
        jsPlumb.detachEveryConnection();
        scope.schema = [];
        scope.library = [];
        scope.addModuleToLibrary("Start", "description",0,0,"http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Cut-icon.png","");
        scope.addModuleToLibrary("Stop", "description",0,0,"http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png","");
        scope.addModuleToLibrary("Actor", "description",0,0,"http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png","");
        scope.addModuleToLibrary("DoWhile", "description",0,0,"http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Add-fav-icon.png",
{
  "Control": {
    "Name": "DoWhile",
    "Icon": "Blah",
    "Variables": [
      {
        "Key": "GUUserID",
        "Value": 4544314512
      },
      {
        "Key": "Username",
        "Value": "shehantis"
      }
    ]
  }
});
        scope.addModuleToLibrary("ForEach<T>", "description",0,0,"http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Chat-icon.png",
{
  "Control": {
    "Name": "ForEach",
    "Icon": "Blah",
    "Variables": [
      {
        "Key": "Password",
        "Value": "fdfgdfgdfg"
      },
      {
        "Key": "Surname",
        "Value": "blah blah"
      }
    ]
  }
});
    };
    

    // add a module to the library
    scope.addModuleToLibrary = function(title, description, posX, posY,icon,variables) {
        console.log("Add module " + title + " to library, at position " + posX + "," + posY+", variables: "+variables);
        var library_id = scope.library_uuid++;
        var schema_id = -1;
        var m = new module(library_id, schema_id, title, description, posX, posY,icon,variables);
        scope.library.push(m);
    };

    // add a module to the schema
    scope.addModuleToSchema = function(library_id, posX, posY) {
        console.log("Add module " + title + " to schema, at position " + posX + "," + posY);
        var schema_id = scope.schema_uuid++;
        var title = "";
        var description = "Likewise unknown";
        var icon = "";
        var variables = "";
        for (var i = 0; i < scope.library.length; i++) {
            if (scope.library[i].library_id == library_id) {
                title = scope.library[i].title;
                description = scope.library[i].description;
                icon = scope.library[i].icon;
                variables = scope.library[i].variables; console.log("Selected control variables : " + variables);
            }
        }
        var m = new module(library_id, schema_id, title, description, posX, posY,icon,variables);
        scope.schema.push(m);
    };

    scope.removeState = function(schema_id) {
        console.log("Remove state " + schema_id + " in array of length " + scope.schema.length);
        for (var i = 0; i < scope.schema.length; i++) {
            // compare in non-strict manner
            if (scope.schema[i].schema_id == schema_id) {
                console.log("Remove state at position " + i);
                scope.schema.splice(i, 1);
            }
        }
    };

    scope.init = function() {
        jsPlumb.bind("ready", function() {
            console.log("Set up jsPlumb listeners (should be only done once)");
            jsPlumb.bind("connection", function (info) {
                scope.$apply(function () {
                    console.log("Possibility to push connection into array");
                });
            });
        });
    }

    /***************************************************************************************/
    /***************************************************************************************/

	scope.activitylist =[
        {name:'Jani',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Hege',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Kai',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Bali',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Baila',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Sindu',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Gamuda',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Remix',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Karala',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Patta',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Pata pata',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Hari',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'},
        {name:'Wage',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini-2/48/Data-icon.png'}
    ];

    scope.controlflowlist =[
        {library_id:0,name:'DoWhile',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Add-fav-icon.png', variables :[{"variable 1":"value 1"}]},
        {library_id:1,name:'ForEach<T>',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Chat-icon.png', variables :[{"variable 1":"value 1"}]},
        {library_id:2,name:'If',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Cut-icon.png'},
        {library_id:3,name:'Parallel',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'},
        {library_id:4,name:'ParallelForEach<T>',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'},
        {library_id:5,name:'Pick',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'},
        {library_id:6,name:'PickBranch',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'},
        {library_id:7,name:'Sequence',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'},
        {library_id:8,name:'Switch<T>',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'},
        {library_id:9,name:'While',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Faq-icon.png'}
    ];

    scope.flowchartlist =[
        {name:'Flowchart',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Add-fav-icon.png'},
        {name:'FlowDecision',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Chat-icon.png'},
        {name:'FlowSwitch<T>',icon:'http://icons.iconarchive.com/icons/custom-icon-design/mini/48/Cut-icon.png'}
    ];

    scope.ZoomIn = function () {  
		var ZoomInValue = parseInt(document.getElementById("container").style.zoom) + 10 + '%'  
		document.getElementById("container").style.zoom = ZoomInValue;  
		return false;  
	}  

	scope.ZoomOut = function () {  
		var ZoomOutValue = parseInt(document.getElementById("container").style.zoom) - 10 + '%'  
		document.getElementById("container").style.zoom = ZoomOutValue;  
		return false;  
	}

	scope.Zoomorg = function () {  
		var ZoomOutValue = parseInt(100) + '%'  
		document.getElementById("container").style.zoom = ZoomOutValue;  
		return false;  
	}

	scope.openBox = function(){
		$("#toolboxControl").css("left","0px");
		$("#openbox").css("display","none");
		$("#closebox").css("display","block");
	}

	scope.closeBox = function(){
		$("#toolboxControl").css("left","-250px");
		$("#closebox").css("display","none");
		$("#openbox").css("display","block");
	}

    scope.vopenBox = function(){
        $("#variablepanel").css("right","0px");
        $("#vopenbox").css("display","none");
        $("#vclosebox").css("display","block");
    }

    scope.vcloseBox = function(){
        $("#variablepanel").css("right","-250px");
        $("#vclosebox").css("display","none");
        $("#vopenbox").css("display","block");
    }

}]);

// this runs on the app load. this will run the redraw method which will draw the required items in the toolbox

app.directive('postRender', [ '$timeout', function($timeout) {
    var def = {
            restrict : 'A', 
            terminal : true,
            transclude : true,
            link : function(scope, element, attrs) {
                $timeout(scope.redraw, 0);  //Calling a scoped method
            }
    };
    return def;
}]);


//directives link user interactions with scope behaviours
//now we extend html with <div plumb-item>, we can define a template <> to replace it with "proper" html, or we can 
//replace it with something more sophisticated, e.g. setting jsPlumb arguments and attach it to a double-click 
//event
app.directive('plumbItem', function() {
    return {
        replace: true,
        controller: 'mainController',
        link: function (scope, element, attrs) {
            console.log("Add plumbing for the 'item' element");

            jsPlumb.makeTarget(element, {
                endpoint:"Blank",
                anchor:[ "Perimeter", { shape:"Square", anchorCount:8 }],
                connectorOverlays:[ 
                    [ "Arrow", { width:30, length:30, location:1, id:"arrow" } ]
                ]
            });
            jsPlumb.draggable(element, {
                containment: 'parent'
            });

            // this should actually done by a AngularJS template and subsequently a controller attached to the dbl-click event
            element.bind('dblclick', function(e) {
                jsPlumb.detachAllConnections($(this));
                $(this).remove();
                // stop event propagation, so it does not directly generate a new state
                e.stopPropagation();
                //we need the scope of the parent, here assuming <plumb-item> is part of the <plumbApp>         
                scope.$parent.removeState(attrs.identifier);
                scope.$parent.$digest();
            });

        }
    };
});

//
// This directive should allow an element to be dragged onto the main canvas. Then after it is dropped, it should be
// painted again on its original position, and the full module should be displayed on the dragged to location.
//
app.directive('plumbMenuItem', function() {
    return {
        replace: true,
        controller: 'mainController',
        link: function (scope, element, attrs) {
            console.log("Add plumbing for the 'menu-item' element");

            // jsPlumb uses the containment from the underlying library, in our case that is jQuery.
            jsPlumb.draggable(element, {
                containment: element.parent().parent()
            });
        }
    };
});

app.directive('plumbConnect', function() {
    return {
        replace: true,
        link: function (scope, element, attrs) {
            console.log("Add plumbing for the 'connect' element");

            jsPlumb.makeSource(element, {
                parent: $(element).parent(),
//              anchor: 'Continuous',
                endpoint:"Blank",
                anchor:[ "Perimeter", { shape:"Square", anchorCount:8 }],
                connectorOverlays:[ 
                [ "Arrow", { width:15, length:15, location:1, id:"arrow" } ]
                ]
            });
        }
    };
});

app.directive('droppable', function($compile) {
    return {
        restrict: 'A',
        link: function(scope, element, attrs){
            

            element.droppable({
                drop:function(event,ui) {
                    console.log("Make this element droppable");
                    // angular uses angular.element to get jQuery element, subsequently data() of jQuery is used to get
                    // the data-identifier attribute
                    var dragIndex = angular.element(ui.draggable).data('identifier'),
                    dragEl = angular.element(ui.draggable),
                    dropEl = angular.element(this);

                    // if dragged item has class menu-item and dropped div has class drop-container, add module 
                    if (dragEl.hasClass('menu-item') && dropEl.hasClass('drop-container')) {
                        console.log("Drag event on " + dragIndex);
                        var x = event.pageX - scope.module_css.width / 2;
                        var y = event.pageY - scope.module_css.height / 2;
                        //var x = e.pageX - $(document).scrollLeft() - $('#container').offset().left;
                        //var y = e.pageY - $(document).scrollTop() - $('#container').offset().top;
                        //alert('x='+x+' y='+y);
                        scope.addModuleToSchema(dragIndex, event.pageX, event.pageY);
                    }

                    scope.$apply();
                }
            });
        }
    };
});

app.directive('draggable', function() {
    return {
        // A = attribute, E = Element, C = Class and M = HTML Comment
        restrict:'A',
        //The link function is responsible for registering DOM listeners as well as updating the DOM.
        link: function(scope, element, attrs) {
            console.log("Let draggable item snap back to previous position");
            element.draggable({
                // let it go back to its original position
                revert:true,
            });
        }
    };
});