import { useEffect, useState } from "react";
import "./App.css";

function App() {
  const [stockData, setStockData] = useState<any>({});

  useEffect(() => {
    const es = new EventSource("http://localhost:8080/stock-events");

    es.onopen = () => console.log(">>> Connection opened!");

    es.onerror = (e) => console.log("ERROR!", e);

    es.onmessage = (e) => {
      console.log(e.data, "e.data");

      setStockData(JSON.parse(e.data));
    };

    // Whenever we're done with the data stream we must close the connection
    return () => es.close();
  }, []);

  return (
    <div>
      <p>Ticker: {stockData.symbol}</p>
      <p>Price: {stockData.price}</p>
      <p>Open: {stockData.open}</p>
      <p>High: {stockData.high}</p>
      <p>Low: {stockData.low}</p>
      <p>PrevClose: {stockData.prevClose}</p>
    </div>
  );
}

export default App;
