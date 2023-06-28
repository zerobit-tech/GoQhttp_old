var socket = null;


// tell Ws that client is leaving
window.onbeforeunload = function () {
                 return
                console.log("Leaving"); 
                let jsonData = {};////WsClientPayload
                jsonData["action"] = "left"; 
                socket.send(JSON.stringify(jsonData)) // send left action to web socket
            }



$(document).ready(function () {
         return
        
        //socket = new WebSocket("ws://127.0.0.1:4000/ws/notification");
        socket = new ReconnectingWebSocket(websocketurl, null, {debug: true, reconectInterval: 3000});

        socket.onopen = () =>{
            console.log("Websocket opened.....")
            let jsonData = {};////WsClientPayload

            jsonData["action"] = "getgraphdata"; 
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
            //console.log(msg)
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
                        if (typeof dashboardgraph !== null) {
                            console.log("updating graph")
                            //dashboardgraph.data.datasets=data.data
                           Plotly.newPlot('dashboardgraph', data.data);
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
                            data.data.Calltime]).draw(false);
                        }
                    break;

             }

        }





})