'use strict';
var uri = "http://192.168.2.40"
var port ="3048"

function getURI() {
        return uri+":"+port
}

var AuthApp = angular.module('AuthApp', ['ngRoute','uiMicrokernel']);

AuthApp.config(function($routeProvider){
		$routeProvider
			.when('/',{
				templateUrl: 'partials/dashboard.html',
				controller: 'dashboardCtrl',
			})
			.when('/dashboard/:file',{
				templateUrl: 'partials/dashboard.html',
				controller: 'dashboardCtrl'
			})
			.when('/config',{
				templateUrl: 'partials/config.html',
				controller: 'configCtrl'
			})
			.when('/config/Auth.config',{
				templateUrl: 'partials/config/Auth.html',//+$routeParams.file.substring(0, $routeParams.file.length - 7)+".html",
				controller: 'authCtrl'
			})
			.when('/config/Terminal.config',{
				templateUrl: 'partials/config/Terminal.html',//+$routeParams.file.substring(0, $routeParams.file.length - 7)+".html",
				controller: 'termCtrl'
			})
			.when('/api',{
				templateUrl: 'partials/apidoc.html',
				controller: 'configCtrl'
			});
	});


AuthApp.controller('configCtrl',['$scope','$http','$routeParams','$v6urls', function($scope, $http, $routeParams,$v6urls){
		$scope.ctr="configCtrl"
		$scope.pageClass = 'page-config';
		$scope.pageName  = 'Configurations';
		console.log($routeParams)
		if($scope.repositoryFiles==null)
		{		
			$http({
				method: 'GET',
				url: getURI()+ '/Config/Files',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				$scope.files = data;
				console.log(data);
			}).error(function(data, status, headers, config){
				console.log(data);
			}); 
		}
		if($routeParams.file!=null)
		{
			$scope.FileName=$routeParams.file;
			$http({
				method: 'GET',
				url: getURI()+'/Config/Get/'+$routeParams.file.substring(0, $routeParams.file.length - 7),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				$scope.FileContent = data;
				console.log(data);
			}).error(function(data, status, headers, config){
				console.log(data);
			}); 
			//$scope.FileContent=$routeParams.page;
		}
	}]);


AuthApp.controller('authCtrl',['$scope','$http','$routeParams', function($scope, $http, $routeParams){
		$scope.ctr="authCtrl"
		$scope.pageClass = 'page-AuthConfig';
		$scope.pageName  = 'Auth Configurations';
		$scope.FileName = "Auth"
			$http({
				method: 'GET',
				url: getURI()+'/Config/Get/'+$scope.FileName,
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				$scope.FileContent = data;
				console.log(data);
			}).error(function(data, status, headers, config){
				console.log(data);
			}); 
	$scope.submit = function() {
		var obj={"FileName":$scope.FileName,"Body":JSON.stringify($scope.FileContent)}
		console.log(JSON.stringify(obj));
		$http({
				method: 'POST',
				url: getURI()+'/Config/Save/',
				data:JSON.stringify(obj),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				//$scope.FileContent = data;
				console.log(data);
			}).error(function(data, status, headers, config){
				console.log(data);
			}); 
    };	
	}]);

AuthApp.controller('termCtrl',['$scope','$http','$routeParams', function($scope, $http, $routeParams){
		$scope.ctr="termCtrl"
		$scope.pageClass = 'page-termConfig';
		$scope.pageName  = 'Terminal Configurations';
		$scope.FileName = "Terminal"
		console.log(getURI()+'/Config/Get/'+$scope.FileName)
			$http({
				method: 'GET',
				url: getURI()+'/Config/Get/'+$scope.FileName,
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				$scope.FileContent = data;
				console.log(data);
			}).error(function(data, status, headers, config){
				console.log(data);
			}); 
	$scope.submit = function() {
		var obj={"FileName":$scope.FileName,"Body":JSON.stringify($scope.FileContent)}
		console.log(JSON.stringify(obj));
		$http({
				method: 'POST',
				url: getURI()+'/Config/Save/',
				data:JSON.stringify(obj),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				//$scope.FileContent = data;
				console.log(data);
			}).error(function(data, status, headers, config){
				console.log(data);
			}); 
    };		
}]);

AuthApp.controller('dashboardCtrl',['$scope','$http','$routeParams','$interval', function($scope, $http, $routeParams,$interval){
		$scope.pageClass = 'page-AuthConfig';
		$scope.pageName  = 'dashboard';
		$scope.ctr="dashboardCtrl";
		$scope.Sucess =[];
        $scope.Error=[];
        $scope.SucessDataRate=[];
        $scope.ErrorDataRate=[];
		$scope.chart = new CanvasJS.Chart("chartContainer", {
        theme: 'theme1',
        title:{
            text: "Number of calls made to the Server"              
        },
        axisY: {
            title: "Number of Calls",
            labelFontSize: 16,
        },
        axisX: {
        	title:"Time",
            labelFontSize: 16,
        },
        data: [              
            {
                type: "line",
                xValueType: "dateTime",
                name: "Sucessful calls",
                dataPoints: $scope.Sucess
            },
            {
                type: "line",
                xValueType: "dateTime",
                name: "Failed calls",
                dataPoints: $scope.Error
            }]
    	});

    	$scope.chart2 = new CanvasJS.Chart("chartTransfer", {
        theme: 'theme1',
        title:{
            text: "Data Transfer Rate"              
        },
        axisY: {
            title: "Data in Bytes",
            labelFontSize: 16,
        },
        axisX: {
        	title:"Time",
            labelFontSize: 16,
        },
        data: [              
            {
                type: "line",
                xValueType: "dateTime",
                name: "Sucessful calls",
                dataPoints: $scope.SucessDataRate
            },
            {
                type: "line",
                xValueType: "dateTime",
                name: "Failed calls",
                dataPoints: $scope.ErrorDataRate
            }]
    	});
    $scope.OldSucessVal=0;
    $scope.OldFailedVal=0;

	$scope.UpdateChart=function(){
		var s =new Date()
    	$http({
				method: 'GET',
				url: getURI()+'/stat/GetStatus/Success',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				if($scope.DataSucess!=null){
					$scope.Sucess.push({x:s.getTime(),y:data.NumberOfCalls- $scope.DataSucess.NumberOfCalls})
					$scope.SucessDataRate.push({x:s.getTime(),y:data.TotalSize- $scope.DataSucess.TotalSize})		
				}else{
					$scope.Sucess.push({x:s.getTime(),y:0})
					$scope.SucessDataRate.push({x:s.getTime(),y:0})
				}
				$scope.DataSucess=data	
				$scope.OldSucessVal=data.NumberOfCalls
			}).error(function(data, status, headers, config){
				console.log(data);
			});
		$http({
				method: 'GET',
				url: getURI()+'/stat/GetStatus/Error',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'}
			}).success(function(data, status, headers, config){
				if($scope.DataError!=null){
					$scope.Error.push({x:s.getTime(),y:data.NumberOfCalls- $scope.DataError.NumberOfCalls})
					$scope.ErrorDataRate.push({x:s.getTime(),y:data.TotalSize- $scope.DataError.TotalSize})
				}else{
					$scope.Error.push({x:s.getTime(),y:0}) 
					$scope.ErrorDataRate.push({x:s.getTime(),y:0}) 					
				}
				$scope.DataError=data
				$scope.OldFailedVal=data.NumberOfCalls
			}).error(function(data, status, headers, config){
				console.log(data);
			});
		$scope.chart.render();
		$scope.chart2.render();
	}
    $scope.UpdateChart();
    $interval($scope.UpdateChart,10000);
    $scope.changeChartType = function(chartType) {
        $scope.chart.options.data[0].type = chartType;
        $scope.chart.render(); 
    }	
}]);

