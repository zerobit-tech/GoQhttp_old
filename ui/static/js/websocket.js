var socket = null;


// tell Ws that client is leaving
window.onbeforeunload = function () {
                  
                console.log("Leaving"); 
                let jsonData = {};////WsClientPayload
                jsonData["action"] = "left"; 
                socket.send(JSON.stringify(jsonData)) // send left action to web socket
            }



$(document).ready(function () {

        var graphUpdateCalls = 0
          
        
        //socket = new WebSocket("ws://127.0.0.1:4000/ws/notification");
        socket = new ReconnectingWebSocket(websocketurl, null, {debug: true, reconectInterval: 3000});

        socket.onopen = () =>{
            console.log("Websocket opened.....")
            let jsonData = {};////WsClientPayload

            jsonData["action"] = "getgraphdata"; 
            socket.send(JSON.stringify(jsonData)); // send left action to websocket
        
            jsonData["action"] = "getgraphstats"; 
            socket.send(JSON.stringify(jsonData)); // send left action to websocket
        
        
        }


    // to send data to websocket
    //socket.send(JSON.stringify(jsonData));

        socket.onclose = () => {
            console.log("connection closed");
            let jsonData = {};////WsClientPayload
            jsonData["action"] = "left"; 
            socket.send(JSON.stringify(jsonData)) // send left action to web socket
        }


        socket.onerror = error => {
                    console.log("there was an error");
               
                }
        socket.onmessage = msg => {
            //console.log("msg.data")
            //console.log(msg.data)
            var data = {}; // WsNotification
            try {
                data = JSON.parse(msg.data);
            } catch (e) {
                //alert(e)
                return false;
            }

            switch (data.action) {
                case "notification":
                       // notify(data.message,data.messagetype)

                        Swal.fire({
                            position: 'top-end',
                            icon: data.messagetype,
                            text: data.message,
                            showConfirmButton: false,
                            timer: 5000
                            })
                            break;
                case "ping":
                        var jsonData = {};////WsClientPayload
                        jsonData["action"] = "pong"; 
                        socket.send(JSON.stringify(jsonData)); // send left action to web socket
                        break;

            
                case "graphdata":
                    
                        var dashboardgraph = document.getElementById("dashboardgraph")
                        if (typeof dashboardgraph !== null && typeof autorefdashboard !== 'undefined'){
                            //console.log("autorefdashboard",autorefdashboard)
                            if (autorefdashboard) {
                                console.log("updating graph")
                                //dashboardgraph.data.datasets=data.data
                               Plotly.newPlot('dashboardgraph', data.data);
                               graphUpdateCalls = graphUpdateCalls + 1
                            }
                           
                        }

                        if (graphUpdateCalls > 500){
                            location.reload()
                        }
                        break;

                case "graphtablercd":


                       
                if (typeof dashboardlisttable !== 'undefined') {
                       
                        var logUrl = "<a href='"+data.data.LogUrl+"'>"+data.data.Requestid+"</a>"
                        var spUrl = "<a href='"+data.data.SpUrl+"'>"+data.data.SpName+"</a>"

                        dashboardlisttable.row.add(
                            [
                                logUrl,
                                spUrl,
                                data.data.Httpcode,
                            data.data.Responsetime,
                            data.data.SPResponsetime,
                            data.data.Calltime]).draw(false);
                        }
                    break;




                case "graphstats":

                var dashboardgraph = document.getElementById("dashboardgraph")
                if (typeof autorefdashboard === 'undefined' || !autorefdashboard){
                    return
                }
                   //  console.log(data.data)
                    if (typeof graphfooter_vue !== null && typeof graphfooter_vue !== 'undefined'){
                        // alert(graphfooter_vue.http100Count)
                         graphfooter_vue.http100Count = data.data.http100count
                         graphfooter_vue.http100Percent=data.data.http100percent

                         graphfooter_vue.http200Count = data.data.http200count
                         graphfooter_vue.http200Percent=data.data.http200percent

                         graphfooter_vue.http300Count = data.data.http300count
                         graphfooter_vue.http300Percent=data.data.http300percent

                         graphfooter_vue.http400Count = data.data.http400count
                         graphfooter_vue.http400Percent=data.data.http400percent

                         graphfooter_vue.http500Count = data.data.http500count
                         graphfooter_vue.http500Percent=data.data.http500percent
                         graphUpdateCalls = graphUpdateCalls + 1

                     }

                     if (typeof graphtableheader !== null && typeof graphtableheader !== 'undefined'){
                        graphtableheader.avgRspTime =data.data.avgrestime
                        graphtableheader.maxRspTime = data.data.maxrestime

                        graphtableheader.avgDBTime = data.data.avgdbtime
                        graphtableheader.maxDBTime =data.data.maxdbtime

                     }
                     if (graphUpdateCalls > 500){
                        location.reload()
                    }


                break;


             }

        }





})