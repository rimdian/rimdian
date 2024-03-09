const AlphaSort = (sortKey: string) => {
  return (a: any, b: any) => {
    if (a[sortKey] < b[sortKey]) {
      return -1
    }
    if (a[sortKey] > b[sortKey]) {
      return 1
    }
    return 0
  }
}

export default AlphaSort
