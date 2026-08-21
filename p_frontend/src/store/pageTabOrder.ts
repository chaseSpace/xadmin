import type { AppTabItem, AppTabKey } from './pageTabs'

export function reorderTabs(
  tabs: AppTabItem[],
  sourceKey: AppTabKey,
  targetKey: AppTabKey,
): AppTabItem[] {
  if (sourceKey === targetKey) return tabs

  const sourceIndex = tabs.findIndex((item) => item.key === sourceKey)
  const targetIndex = tabs.findIndex((item) => item.key === targetKey)
  if (sourceIndex < 0 || targetIndex < 0) return tabs

  const nextTabs = [...tabs]
  const [sourceTab] = nextTabs.splice(sourceIndex, 1)
  nextTabs.splice(targetIndex, 0, sourceTab)

  const orderChanged = nextTabs.some((item, index) => item.key !== tabs[index]?.key)
  return orderChanged ? nextTabs : tabs
}
