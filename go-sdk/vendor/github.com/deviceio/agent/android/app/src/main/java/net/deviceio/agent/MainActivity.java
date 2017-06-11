package net.deviceio.agent;

import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ToggleButton;

import java.util.UUID;

/**
 * Entry activity
 */
public class MainActivity extends AppCompatActivity {
    /**
     * Holds the active TransportService intent
     */
    private Intent transportServiceIntent;

    /**
     * called when the activity is created
     * @param savedInstanceState
     */
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        SharedPreferences sharedPref = this.getPreferences(Context.MODE_PRIVATE);

        if (!sharedPref.contains("agent_id")) {
            sharedPref.edit().putString("agent_id", UUID.randomUUID().toString()).commit();
        }

        if (sharedPref.contains("hub_host")) {
            EditText hub_host = (EditText) findViewById(R.id.hub_host);
            hub_host.setText(sharedPref.getString("hub_host", ""));
        }

        if (sharedPref.contains("hub_port")) {
            EditText hub_port = (EditText) findViewById(R.id.hub_port);
            hub_port.setText(sharedPref.getString("hub_port", ""));
        }
    }

    /**
     * Called when the agent state button is toggled in the main activity ui
     * @param v
     */
    public void onAgentStateClick(View v) {
        SharedPreferences sharedPref = this.getPreferences(Context.MODE_PRIVATE);
        final ToggleButton agent_state = (ToggleButton) findViewById(R.id.agent_state);
        if (agent_state.isChecked()) {
            EditText hub_host = (EditText) findViewById(R.id.hub_host);
            EditText hub_port = (EditText) findViewById(R.id.hub_port);

            sharedPref.edit().putString("hub_host", hub_host.getText().toString()).commit();
            sharedPref.edit().putString("hub_port", hub_port.getText().toString()).commit();

            TransportService.agentId = sharedPref.getString("agent_id", "android-invalid-id");
            TransportService.hubHost = hub_host.getText().toString();
            TransportService.hubPort = Integer.parseInt(hub_port.getText().toString());

            this.transportServiceIntent = new Intent(v.getContext(), TransportService.class);
            v.getContext().startService(this.transportServiceIntent);
        } else {
            v.getContext().stopService(this.transportServiceIntent);
        }
    }
}
