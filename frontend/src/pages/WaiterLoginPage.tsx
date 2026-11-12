import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { requestWaiterOtp, verifyWaiterOtp } from "../api/auth";

export default function WaiterLoginPage() {
  const navigate = useNavigate();
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [step, setStep] = useState<"phone" | "otp">("phone");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function sendOtp() {
    setLoading(true);
    setError("");
    try {
      await requestWaiterOtp(phone);
      setStep("otp");
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "Unable to send verification code.");
    } finally {
      setLoading(false);
    }
  }

  async function verify() {
    setLoading(true);
    setError("");
    try {
      await verifyWaiterOtp(phone, otp);
      navigate("/waiter");
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "Unable to verify the code.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-[#fffaf5] px-4 py-10 text-slate-950">
      <div className="mx-auto max-w-lg rounded-[2rem] border border-slate-200 bg-white p-8 shadow-xl">
        <p className="text-sm font-black uppercase tracking-[0.3em] text-orange-600">Waiter access</p>
        <h1 className="mt-3 text-4xl font-black">Staff sign in</h1>
        <p className="mt-3 text-slate-600">Use the phone number approved by an administrator.</p>
        {step === "phone" ? (
          <>
            <label className="mt-8 block text-sm font-bold">Mobile number</label>
            <input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="9876543210" inputMode="tel" className="mt-2 w-full rounded-2xl border border-slate-200 px-4 py-4 text-lg outline-none focus:ring-2 focus:ring-orange-300" />
            <button type="button" disabled={loading} onClick={() => void sendOtp()} className="mt-5 w-full rounded-2xl bg-slate-950 px-4 py-4 font-black text-white disabled:opacity-50">{loading ? "Sending..." : "Send OTP"}</button>
          </>
        ) : (
          <>
            <label className="mt-8 block text-sm font-bold">Six-digit OTP</label>
            <input value={otp} onChange={(event) => setOtp(event.target.value.replace(/\D/g, "").slice(0, 6))} placeholder="123456" inputMode="numeric" autoFocus className="mt-2 w-full rounded-2xl border border-slate-200 px-4 py-4 text-center text-2xl tracking-[0.5em] outline-none focus:ring-2 focus:ring-orange-300" />
            <button type="button" disabled={loading || otp.length !== 6} onClick={() => void verify()} className="mt-5 w-full rounded-2xl bg-orange-600 px-4 py-4 font-black text-white disabled:opacity-50">{loading ? "Verifying..." : "Enter waiter dashboard"}</button>
            <button type="button" onClick={() => setStep("phone")} className="mt-3 w-full rounded-2xl border border-slate-200 px-4 py-3 font-bold text-slate-700">Change number</button>
          </>
        )}
        {error ? <p role="alert" className="mt-4 rounded-2xl bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700">{error}</p> : null}
        <button type="button" onClick={() => navigate("/")} className="mt-6 text-sm font-bold text-slate-500 hover:text-orange-600">Back to restaurant</button>
      </div>
    </div>
  );
}
