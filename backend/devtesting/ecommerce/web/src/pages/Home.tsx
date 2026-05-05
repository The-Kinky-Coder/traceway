type Props = { onShop: () => void };

export default function Home({ onShop }: Props) {
  return (
    <div>
      <h1>Quietly considered everyday goods.</h1>
      <p className="lead">
        Loomstead makes a small line of clothing and accessories from natural fibers and traditional makers. New
        drops monthly. Returns within 30 days, no questions asked.
      </p>
      <button onClick={onShop}>Shop the catalog →</button>
    </div>
  );
}
