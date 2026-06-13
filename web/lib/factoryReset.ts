/**
 * Single source of truth for factory-reset disclosure copy (REQ-017 BR-2)
 * and the confirmation token the API requires (WI-15).
 * Every category named in the business rule must appear here so that the
 * confirm dialog, the summary paragraph, and any future surface stay in sync.
 */

/**
 * The phrase the user must type (and the API must receive) to authorise a
 * factory reset. Kept here so the UI and any future callers stay in sync.
 */
export const FACTORY_RESET_CONFIRM_PHRASE = "FACTORY RESET";

export const FACTORY_RESET_DISCLOSURE =
  "Permanently removes every model you uploaded, every scene, all registered " +
  "devices, and every routine you created (Python and shape-animation). Any " +
  "in-flight routine runs will be stopped. After the reset finishes, you " +
  "will only see the three default sample models and the three default sample " +
  "Python routines.";

export const POST_RESET_FLASH =
  "All data was reset. The three default sample models and three default " +
  "sample Python routines were restored.";
