interface Props {
  error: Error;
  resetErrorBoundary: () => void;
}

const Fallback = ({ error, resetErrorBoundary }: Props) => {
  // Call resetErrorBoundary() to reset the error boundary and retry the render.
  console.debug({ error });
  return (
    <div role="alert" style={{ padding: 24 }}>
      <p>Something went wrong:</p>
      <pre style={{ color: "red" }}>{error.message}</pre>
      <pre>{error.stack}</pre>
      <button onClick={resetErrorBoundary}>Reset</button>
    </div>
  );
};

export default Fallback;
