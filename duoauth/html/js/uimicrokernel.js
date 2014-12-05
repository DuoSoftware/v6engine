var microKernelModule = angular.module('uiMicrokernel', []);

microKernelModule.factory('$objectstore', function($http, $v6urls) {
  
	function Requestor(_namespace,_class,_token){

		var namespace = _namespace;
		var cls = _class;
		var token = _token;
		var onGetOne;
		var onGetMany;
		var onComplete;
		var onError;

		function insertLogic(data,parameters){

			var mainObject = null;
			if(angular.isArray(data))
				mainObject = {Parameters : parameters, Objects : data};
			else
				mainObject = {Parameters : parameters, Object : data};


			$http.post($v6urls.objectStore + '/' + namespace + '/' + cls,mainObject, {headers:{"securityToken" : "123"}}).
			  success(function(data, status, headers, config) {
			  	if (onComplete)
			  		onComplete(data);				  	
			  }).
			  error(function(data, status, headers, config) {
			  	if (onError)
			  		onError()

			  	if (onComplete){
			  		if (data)
			  			onComplete(data);
			  		else
			  			onComplete({isSuccess:false, message:"Unknown Error!!!"});
			  	}
			  });
		}

		return {
			getByKeyword: function(keyword,parameters){
				$http.get($v6urls.objectStore + '/' + namespace + '/' + cls + '?keyword=' + keyword,{headers:{"securityToken" : "123"}}).
				  success(function(data, status, headers, config) {
				  	if (onGetMany)
				  		onGetMany(data);				  	
				  }).
				  error(function(data, status, headers, config) {
				  	if (onError)
				  		onError()

				  	if (onGetMany)
				  		onGetMany();
				  });
			},
			getByKey: function(key){
				$http.get($v6urls.objectStore + '/' + namespace + '/' + cls + '/' + key,{headers:{"securityToken" : "123"}}).
				  success(function(data, status, headers, config) {
				  	if (onGetOne)
				  		onGetOne(data);				  	
				  }).
				  error(function(data, status, headers, config) {
				  	if (onError)
				  		onError()

				  	if (onGetOne)
				  		onGetOne();
				  });
			},
			getAll: function(parameters){
				
			},
			getByFiltering: function(filter,parameters){
				//,"Content-Type":"application/json"
				$http.post($v6urls.objectStore + '/' + namespace + '/' + cls ,{"Query" : {"Type" : "", "Parameters": filter}}, {headers:{"securityToken" : "123"}}).
				  success(function(data, status, headers, config) {
				  	if (onGetMany)
				  		onGetMany(data);				  	
				  }).
				  error(function(data, status, headers, config) {
				  	if (onError)
				  		onError()

				  	if (onGetMany)
				  		onGetMany();
				  });
			},
			insert: function(data,parameters){
				insertLogic(data,parameters);
			},
			update: function(data,parameters){
				insertLogic(data,parameters);
			},
			delete: function(data,parameters){

			},
			onGetOne: function(func){ onGetOne = func },
			onGetMany:function(func){ onGetMany = func },
			onComplete:function(func){ onComplete = func },
			onError: function(func){ onError = func}
		}
	}


	return {
  		getRequestor:function(namespace,cls,token){
  			var req = new Requestor(namespace,cls,token);
  			return req;
  		}
  	}
});


microKernelModule.factory('$auth', function($http, $v6urls) {
 
 	var sessionInfo;
 	var securityToken;
	var onLoggedInResultEvent;

	function login(username, password,domain){
		var loginResult = {isSuccess:true, message:"Success", securityToken:"", details:{}};

		$http.get($v6urls.auth + "/Login/" + username +"/" + password + "/" + domain).
		  success(function(data, status, headers, config) {
		  	loginResult.details = data;
		  	loginResult.securityToken = data.SecurityToken;
		  	
		  	sessionInfo = data;
		  	securityToken = data.SecurityToken;

		  	if (onLoggedInResultEvent)
		  		onLoggedInResultEvent(loginResult);				  	
		  }).
		  error(function(data, status, headers, config) {
		  	loginResult.isSuccess = false;
		  	loginResult.message = data;
		  	if (onLoggedInResultEvent)
				onLoggedInResultEvent(loginResult);
		  });

		
	}

	return {
  		login: function(username,password, domain){
  			login(username, password, domain)
  		},
  		logout: function(securityToken){
  			var req = new Requestor(namespace,cls,token);
  			return req;
  		},
  		onLoginResult: function(func){
  			onLoggedInResultEvent = func;
  		}

  	}
});

microKernelModule.factory('$fws', function($rootScope, $v6urls) {
    var socket = io.connect($v6urls.fws + "/");
    return {
        on: function(eventName, callback) {
            socket.on(eventName, function() {
                var args = arguments;
                $rootScope.$apply(function() {
                    callback.apply(socket, args);
                });
            });
        },
        emit: function(eventName, data, callback) {
            socket.emit(eventName, data, function() {
                var args = arguments;
                $rootScope.$apply(function() {
                    if (callback) {
                        callback.apply(socket, args);
                    }
                });
            });
        }
    };
});

microKernelModule.factory('$backdoor', function() {
   
   	var logLines = [];
	var onItemAdded;

	function timeStamp() {
		var now = new Date();
		var date = [ now.getMonth() + 1, now.getDate(), now.getFullYear() ];
		var time = [ now.getHours(), now.getMinutes(), now.getSeconds() ];
		 
		var suffix = ( time[0] < 12 ) ? "AM" : "PM";
		 
		time[0] = ( time[0] < 12 ) ? time[0] : time[0] - 12;
		 
		time[0] = time[0] || 12;
		 
		for ( var i = 1; i < 3; i++ )
			if ( time[i] < 10 )
				time[i] = "0" + time[i];
		 
		return date.join("/") + " " + time.join(":") + " " + suffix;
	} 

    return {
		log: function(data){
			var newLine = timeStamp() + "           " +  data;

			logLines.push(newLine);

			if (onItemAdded){
				onItemAdded(newLine, logLines);
			}
		},
		onItemAdded: function(func){
			onItemAdded = func;
		}
    };
});

microKernelModule.factory('$v6urls', function() {
   
	var urls={
		auth:"http://192.168.0.128:3048",
		objectStore:"http://192.168.2.42:3000",
		fws:"http://192.168.2.42:4000"
	};

    return urls;
});