function randomString(length: number) {
  const chars =
    "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
  let result = "";
  for (let i = length; i > 0; --i) {
    result += chars[Math.floor(Math.random() * chars.length)];
  }

  let hash = "";
  for (let i = 0; i < 32; i++) {
    hash += Math.floor(Math.random() * 16).toString(16);
  }
  return hash;
}

export { randomString };
